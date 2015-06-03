package main

import (
	"bufio"
	"communitrix/cmd/cbt"
	"communitrix/cmd/rx"
	"communitrix/cmd/tx"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

var playerUUIDMutex = &sync.Mutex{}
var playerUUID int64 = 0

func NextPlayerUUID() int64 {
	playerUUIDMutex.Lock()
	defer playerUUIDMutex.Unlock()
	playerUUID++
	return playerUUID
}

// Player is the base struct representing connected entities.
type Player struct {
	Mutex    sync.Mutex   // The lock for this player.
	UUID     string       // The player unique identifier on the server.
	Username string       // The username the player has picked.
	Level    int          // This player's level.
	Conn     net.Conn     // The player connection socket itself.
	Send     chan tx.Base // Outbound messages are in a buffered channel.
	Combat   *Combat      // The combat the player is currently in.
}

func (player *Player) AsSendable() *map[string]interface{} {
	return &map[string]interface{}{
		"uuid":     player.UUID,
		"username": player.Username,
		"level":    player.Level,
	}
}

func (cli *Player) IsInCombat() bool {
	cli.Mutex.Lock()
	defer cli.Mutex.Unlock()
	return cli.Combat != nil
}
func (cli *Player) JoinCombat(combat *Combat) {
	cli.Mutex.Lock()
	defer cli.Mutex.Unlock()
	combat.CommandQueue <- cbt.Wrap(cbt.AddPlayer{Player: cli})
	cli.Combat = combat
}
func (cli *Player) LeaveCombat() bool {
	cli.Mutex.Lock()
	defer cli.Mutex.Unlock()
	if cli.Combat != nil {
		cli.Combat.CommandQueue <- cbt.Wrap(cbt.RemovePlayer{Player: cli})
		return true
	}
	return false
}

// NewPlayer is exposed on the Hub class.
func (hub *Hub) NewPlayer(conn net.Conn) *Player {
	return &Player{
		Mutex:  sync.Mutex{},
		UUID:   fmt.Sprintf("CLI%d", NextPlayerUUID()),
		Conn:   conn,
		Send:   make(chan tx.Base, *config.ClientSendBufferSize),
		Combat: nil,
	}
}

// CommandFromPacket processes a JSON-formated payload and attempts to transform it into a hub command.
func (cli *Player) CommandFromPacket(line []byte) *rx.Base {
	// Deserialize the line to a map[string]interface{}.
	var rec map[string]interface{}
	if err := json.Unmarshal(line, &rec); err != nil {
		log.Warning("[comms] Player %s has sent a packet we are unable to unmarshall: %s - %s.", cli.UUID, err, line)
		return nil
	}

	typ := rec["type"].(string)
	switch typ {
	// Those commands need to pass through the hub.
	case "Register":
		return rx.Wrap(cli, rx.Register{Username: rec["username"].(string)})
	case "CombatList":
		return rx.Wrap(cli, rx.CombatList{})
	case "CombatJoin":
		return rx.Wrap(cli, rx.CombatJoin{UUID: rec["uuid"].(string)})

	// User wants to play his turn.
	case "CombatPlayTurn":
		if !cli.IsInCombat() {
			log.Warning("Player %s tried to play a turn in a combat he is not participating to.")
			cli.Send <- tx.Wrap(tx.Error{
				Code:   422,
				Reason: "You cannot send turns to combats you're not participating to.",
			})
		}
		// TODO: Play turn.
	case "CombatLeave":
		if !cli.LeaveCombat() {
			log.Warning("Player %s requested to leave combat, but he is not in one.", cli.UUID)
			cli.Send <- tx.Wrap(tx.Error{
				Code:   422,
				Reason: "You cannot leave a combat while not participating one.",
			})
		}

	default:
		log.Warning("Player %s sent an unhandled command type: %s.", cli.UUID, rec)
		cli.Send <- tx.Wrap(tx.Error{
			Code:   422,
			Reason: "The command you sent could not be understood by the server.",
		})
	}
	return nil
}

// ReadLoop pumps messages from the player to the hub.
func (cli *Player) ReadLoop(cq chan rx.Base) {
	// Whenever the read loop exits, unregister the player from the hub and close the connection.
	defer func() {
		log.Info("Player %s will be disconnected.", cli.UUID)
		// Tell hub about the disconnect.
		cq <- *rx.Wrap(cli, rx.Unregister{})
	}()

	// Loop for every JSON packet received.
	for {
		// Try to read a JSON message from the connection, and check for failure.
		reader := bufio.NewReader(cli.Conn)
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		if cmd := cli.CommandFromPacket(line); cmd != nil {
			cq <- *cmd
		}
	}
}

// WriteLoop pumps messages from the hub to the player.
func (cli *Player) WriteLoop() {
	ticker := time.NewTicker(time.Second * 1)
	// Whenever this method returns, stop the ping timer for this connection and close the it.
	defer func() {
		ticker.Stop()
		cli.Conn.Close()
	}()

	// Loop until data is ready to be sent.
	json := json.NewEncoder(cli.Conn)
	for {
		select {
		// Data is ready to be sent on channel.
		case message, ok := <-cli.Send:
			// An error occured while retrieving the queued message.
			if !ok {
				return
			}
			if _, err := fmt.Fprintf(cli.Conn, "%s\r", message.Type); err != nil {
				log.Warning("Failed to send next packet type to player %s: %s", cli.UUID, err)
				return
			}
			// Try to send the message, and handle failure.
			if _, err := fmt.Fprintf(cli.Conn, "%s\n", json.Encode(message.Command)); err != nil {
				log.Warning("Failed to send packet to player %s: %s", cli.UUID, err)
				return
			}
		case <-ticker.C:
		}
	}
}

func random(min, max int) int {
	return rand.Intn(max-min) + min
}
