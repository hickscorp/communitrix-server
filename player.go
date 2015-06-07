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
	mutex        sync.Mutex       // The lock for this player.
	uuid         string           // The player unique identifier on the server.
	username     string           // The username the player has picked.
	level        int              // This player's level.
	connection   net.Conn         // The player connection socket itself.
	commandQueue chan interface{} // Outbound messages are in a buffered channel.
	combat       i.Combat         // The combat the player is currently in.
}

func (player *Player) UUID() string                     { return player.uuid }
func (player *Player) Username() string                 { return player.username }
func (player *Player) SetUsername(username string)      { player.username = username }
func (player *Player) Level() int                       { return player.level }
func (player *Player) Connection() net.Conn             { return player.connection }
func (player *Player) Combat() i.Combat                 { return player.combat }
func (player *Player) CommandQueue() chan<- interface{} { return player.commandQueue }

func NewPlayer(connection net.Conn) *Player {
	return &Player{
		mutex:        sync.Mutex{},
		uuid:         fmt.Sprintf("CLI%d", NextPlayerUUID()),
		connection:   connection,
		commandQueue: make(chan interface{}, *config.ClientSendBufferSize),
		combat:       nil,
	}
}
func StartNewPlayer(cq chan<- interface{}, connection net.Conn) {
	player := NewPlayer(connection)
	// Send our welcome message.
	player.CommandQueue() <- tx.Wrap(tx.Welcome{Message: "Hi there!"})
	// Start the writing loop thread, then start reading from the connection.
	go player.writeLoop()
	player.readLoop(cq)
}

func (player *Player) AsSendable() util.MapHelper {
	return util.MapHelper{
		"uuid":     player.uuid,
		"username": player.username,
		"level":    player.level,
	}
}

func (player *Player) WhileLocked(do func()) {
	player.mutex.Lock()
	do()
	player.mutex.Unlock()
}
func (player *Player) IsInCombat() bool {
	var ret bool
	player.WhileLocked(func() {
		ret = player.Combat != nil
	})
	return ret
}
func (player *Player) JoinCombat(combat i.Combat) {
	combat.CommandQueue() <- cbt.Wrap(cbt.AddPlayer{Player: player})
	player.WhileLocked(func() {
		player.combat = combat
	})
}
func (player *Player) LeaveCombat() bool {
	if ret := player.IsInCombat(); ret {
		player.combat.CommandQueue() <- cbt.Wrap(cbt.RemovePlayer{Player: player})
		return true
	} else {
		return false
	}
}

// CommandFromPacket processes a JSON-formated payload and attempts to transform it into a hub command.
func (player *Player) CommandFromPacket(line []byte) *rx.Base {
	// Deserialize the line to a util.MapHelper.
	var rec util.MapHelper
	if err := json.Unmarshal(line, &rec); err != nil {
		log.Warning("[comms] Player %s has sent a packet we are unable to unmarshall: %s - %s.", player.uuid, err, line)
		return nil
	}

	typ := rec.String("type")
	switch typ {
	// Those commands need to pass through the hub.
	case "Register":
		return rx.Wrap(player, rx.Register{Username: rec.String("username")})

	// User wants a list of existing combats.
	case "CombatList":
		return rx.Wrap(player, rx.CombatList{})

	// User wants to join the combat.
	case "CombatJoin":
		return rx.Wrap(player, rx.CombatJoin{
			UUID: rec.String("uuid"),
		})

	// User wants to play his turn.
	case "CombatPlayTurn":
		if !player.IsInCombat() {
			log.Warning("Player %s requested to play a turn combat, but he is not in a combat.", player.uuid)
			player.commandQueue <- tx.Wrap(tx.Error{
				Code:   422,
				Reason: "You cannot leave a combat while not participating a combat.",
			})
			break
		}
		player.WhileLocked(func() {
			player.combat.CommandQueue() <- cbt.Wrap(cbt.PlayTurn{
				Player:      player,
				UUID:        rec.String("uuid"),
				Rotation:    logic.NewQuaternionFromMap(rec.Map("rotation")),
				Translation: logic.NewVectorFromMap(rec.Map("translation")),
			})
		})

	// User wants to leave the combat.
	case "CombatLeave":
		if !player.LeaveCombat() {
			log.Warning("Player %s requested to leave combat, but he is not in one.", player.uuid)
			player.commandQueue <- tx.Wrap(tx.Error{
				Code:   422,
				Reason: "You cannot leave a combat while not participating one.",
			})
			break
		}

	default:
		log.Warning("Player %s sent an unhandled command type: %s.", player.uuid, rec)
		player.commandQueue <- tx.Wrap(tx.Error{
			Code:   422,
			Reason: "The command you sent could not be understood by the server.",
		})
		break
	}
	return nil
}

// ReadLoop pumps messages from the player to the hub.
func (player *Player) readLoop(cq chan<- interface{}) {
	// Whenever the read loop exits, unregister the player from the hub and close the connection.
	defer func() {
		// Signal our write queue to exit.
		player.commandQueue <- nil
		// Signal our hub to stop handling this client.
		cq <- *rx.Wrap(player, rx.Unregister{})
	}()
	// Prepare our json reader directly from the connection.
	reader := bufio.NewReader(player.connection)
	// Loop for every JSON packet received.
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		if cmd := player.CommandFromPacket(line); cmd != nil {
			cq <- *cmd
		}
	}
}

// WriteLoop pumps messages from the hub to the player.
func (player *Player) writeLoop() {
	// Loop until data is ready to be sent.
	json := json.NewEncoder(player.connection)
	for {
		select {
		// Data is ready to be sent on channel.
		case iCmd, ok := <-player.commandQueue:
			// An error occured while retrieving the queued message.
			if !ok {
				return
			}

			// Based on the command type...
			switch cmd := iCmd.(type) {
			// This is the way for the hub to interript the write loop.
			case nil:
				return
			// We got ourself a nice command!
			case tx.Base:
				if _, err := fmt.Fprintf(player.connection, "%s\r", cmd.Type); err != nil {
					log.Warning("Failed to send next packet type to player %s: %s", player.uuid, err)
					return
				}
				// Try to send the message, and handle failure.
				if _, err := fmt.Fprintf(player.connection, "%s\n", json.Encode(cmd.Command)); err != nil {
					log.Warning("Failed to send packet to player %s: %s", player.uuid, err)
					return
				}
			}
		case <-time.After(1 * time.Second):
		}
	}
}

func random(min, max int) int {
	return rand.Intn(max-min) + min
}
