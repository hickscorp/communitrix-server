package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gogs.pierreqr.fr/doodloo/communitrix/cmd/cbt"
	"gogs.pierreqr.fr/doodloo/communitrix/cmd/rx"
	"gogs.pierreqr.fr/doodloo/communitrix/cmd/tx"
	"gogs.pierreqr.fr/doodloo/communitrix/i"
	"gogs.pierreqr.fr/doodloo/communitrix/logic"
	"gogs.pierreqr.fr/doodloo/communitrix/util"
	"math/rand"
	"net"
	"sync"
	"time"
)

var (
	playerUUIDMutex       = &sync.Mutex{}
	playerUUID      int64 = 0
)

func NextPlayerUUID() int64 {
	playerUUIDMutex.Lock()
	defer playerUUIDMutex.Unlock()
	playerUUID++
	return playerUUID
}

// Player is the base struct representing connected entities.
type Player struct {
	mutex        sync.Mutex    // The lock for this player.
	uuid         string        // The player unique identifier on the server.
	username     string        // The username the player has picked.
	level        int           // This player's level.
	connection   net.Conn      // The player connection socket itself.
	commandQueue chan *tx.Base // Outbound messages are in a buffered channel.
	exit         chan bool     // Signal exit.
	combat       i.Combat      // The combat the player is currently in.
}

func (this *Player) UUID() string                { return this.uuid }
func (this *Player) Username() string            { return this.username }
func (this *Player) SetUsername(username string) { this.username = username }
func (this *Player) Level() int                  { return this.level }
func (this *Player) Connection() net.Conn        { return this.connection }
func (this *Player) Notify(cmd *tx.Base)         { this.commandQueue <- cmd }
func (this *Player) Combat() i.Combat            { return this.combat }

func NewPlayer(connection net.Conn) *Player {
	return &Player{
		mutex:        sync.Mutex{},
		uuid:         fmt.Sprintf("CLI%d", NextPlayerUUID()),
		connection:   connection,
		commandQueue: make(chan *tx.Base, *config.ClientSendBufferSize),
		exit:         make(chan bool, 1),
		combat:       nil,
	}
}
func StartNewPlayer(hubQueue chan<- *rx.Base, connection net.Conn) {
	player := NewPlayer(connection)
	// Send our welcome message.
	player.commandQueue <- tx.Wrap(tx.Welcome{Message: "Hi there!"})
	// Start the writing loop thread, then start reading from the connection.
	go player.writeLoop()
	player.readLoop(hubQueue)
}

func (this *Player) AsSendable() util.MapHelper {
	return util.MapHelper{
		"uuid":     this.uuid,
		"username": this.username,
		"level":    this.level,
	}
}

func (this *Player) WhileLocked(do func()) {
	this.mutex.Lock()
	do()
	this.mutex.Unlock()
}
func (this *Player) IsInCombat() bool {
	var ret bool
	this.WhileLocked(func() {
		ret = this.Combat != nil
	})
	return ret
}
func (this *Player) JoinCombat(combat i.Combat) {
	combat.Notify(cbt.Wrap(cbt.AddPlayer{Player: this}))
	this.WhileLocked(func() {
		this.combat = combat
	})
}
func (this *Player) LeaveCombat() bool {
	var ret bool
	this.WhileLocked(func() {
		ret := this.combat != nil
		if ret {
			this.combat.Notify(cbt.Wrap(cbt.RemovePlayer{Player: this}))
		}
	})
	return ret
}

// CommandFromPacket processes a JSON-formated payload and attempts to transform it into a hub command.
func (this *Player) CommandFromPacket(line []byte) *rx.Base {
	// Deserialize the line to a util.MapHelper.
	var rec util.MapHelper
	if err := json.Unmarshal(line, &rec); err != nil {
		log.Warning("[comms] Player %s has sent a packet we are unable to unmarshall: %s - %s.", this.uuid, err, line)
		return nil
	}

	typ := rec.String("type")
	switch typ {
	// Those commands need to pass through the hub.
	case "Register":
		return rx.Wrap(this, rx.Register{Username: rec.String("username")})

	// User wants a list of existing combats.
	case "CombatList":
		return rx.Wrap(this, rx.CombatList{})

	// User wants to join the combat.
	case "CombatJoin":
		return rx.Wrap(this, rx.CombatJoin{
			UUID: rec.String("uuid"),
		})

	// User wants to play his turn.
	case "CombatPlayTurn":
		if !this.IsInCombat() {
			log.Warning("Player %s requested to play a turn combat, but he is not in a combat.", this.uuid)
			this.commandQueue <- tx.Wrap(tx.Error{
				Code:   422,
				Reason: "You cannot leave a combat while not participating a combat.",
			})
			break
		}
		this.WhileLocked(func() {
			this.combat.Notify(cbt.Wrap(cbt.PlayTurn{
				Player:      this,
				UUID:        rec.String("uuid"),
				Rotation:    logic.NewQuaternionFromMap(rec.Map("rotation")),
				Translation: logic.NewVectorFromMap(rec.Map("translation")),
			}))
		})

	// User wants to leave the combat.
	case "CombatLeave":
		if !this.LeaveCombat() {
			log.Warning("Player %s requested to leave combat, but he is not in one.", this.uuid)
			this.commandQueue <- tx.Wrap(tx.Error{
				Code:   422,
				Reason: "You cannot leave a combat while not participating one.",
			})
			break
		}

	default:
		log.Warning("Player %s sent an unhandled command type: %s.", this.uuid, rec)
		this.commandQueue <- tx.Wrap(tx.Error{
			Code:   422,
			Reason: "The command you sent could not be understood by the server.",
		})
		break
	}
	return nil
}

// ReadLoop pumps messages from the this to the hub.
func (this *Player) readLoop(hubQueue chan<- *rx.Base) {
	// Whenever the read loop exits, unregister the player from the hub and close the connection.
	defer func() {
		// Signal our write queue to exit.
		this.exit <- true
		// Signal our hub to stop handling this client.
		hubQueue <- rx.Wrap(this, rx.Unregister{})
	}()
	// Prepare our json reader directly from the connection.
	reader := bufio.NewReader(this.connection)
	// Loop for every JSON packet received.
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		if cmd := this.CommandFromPacket(line); cmd != nil {
			hubQueue <- cmd
		}
	}
}

// WriteLoop pumps messages from the hub to the player.
func (this *Player) writeLoop() {
	// Loop until data is ready to be sent.
	json := json.NewEncoder(this.connection)
	for {
		select {
		// Data is ready to be sent on channel.
		case cmd, ok := <-this.commandQueue:
			// An error occured while retrieving the queued message.
			if !ok {
				return
			}

			// We got ourself a nice command!
			if _, err := fmt.Fprintf(this.connection, "%s\r", cmd.Type); err != nil {
				log.Warning("Failed to send next packet type to player %s: %s", this.uuid, err)
				return
			}
			// Try to send the message, and handle failure.
			if _, err := fmt.Fprintf(this.connection, "%s\n", json.Encode(cmd.Command)); err != nil {
				log.Warning("Failed to send packet to player %s: %s", this.uuid, err)
				return
			}
		// This player was asked to exit the loop.
		case <-this.exit:
			return
		case <-time.After(1 * time.Second):
		}
	}
}

func random(min, max int) int {
	return rand.Intn(max-min) + min
}
