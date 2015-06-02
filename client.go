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

var clientUUIDMutex = &sync.Mutex{}
var clientUUID int64 = 0

func NextClientUUID() int64 {
	clientUUIDMutex.Lock()
	defer clientUUIDMutex.Unlock()
	clientUUID++
	return clientUUID
}

// Client is the base struct representing connected entities.
type Client struct {
	Mutex    sync.Mutex   // The lock for this client.
	UUID     string       // The client unique identifier on the server.
	Username string       // The username the client has picked.
	Conn     net.Conn     // The client connection socket itself.
	Send     chan tx.Base // Outbound messages are in a buffered channel.
	Combat   *Combat      // The combat the player is currently in.
}

func (cli *Client) IsInCombat() bool {
	cli.Mutex.Lock()
	defer cli.Mutex.Unlock()
	return cli.Combat != nil
}
func (cli *Client) JoinCombat(combat *Combat) {
	cli.Mutex.Lock()
	defer cli.Mutex.Unlock()
	combat.CommandQueue <- cbt.Wrap(cbt.AddClient{cli})
	cli.Combat = combat
}
func (cli *Client) LeaveCombat() bool {
	cli.Mutex.Lock()
	defer cli.Mutex.Unlock()
	if cli.Combat != nil {
		cli.Combat.CommandQueue <- cbt.Wrap(cbt.RemoveClient{cli})
		return true
	}
	return false
}

// NewClient is exposed on the Hub class.
func (hub *Hub) NewClient(conn net.Conn) *Client {
	return &Client{
		Mutex:  sync.Mutex{},
		UUID:   fmt.Sprintf("CLI%d", NextClientUUID()),
		Conn:   conn,
		Send:   make(chan tx.Base, *config.ClientSendBufferSize),
		Combat: nil,
	}
}

// CommandFromPacket processes a JSON-formated payload and attempts to transform it into a hub command.
func (cli *Client) CommandFromPacket(line []byte) *rx.Base {
	// Deserialize the line to a map[string]interface{}.
	var rec map[string]interface{}
	if err := json.Unmarshal(line, &rec); err != nil {
		log.Warning("[comms] Client %s has sent a packet we are unable to unmarshall: %s - %s.", cli.UUID, err, line)
		return nil
	}

	typ := rec["type"].(string)
	switch typ {
	case "Error":
		return rx.Wrap(cli, rx.Error{rec["code"].(int), rec["reason"].(string)})
	case "Register":
		return rx.Wrap(cli, rx.Register{rec["username"].(string)})
	case "CombatList":
		return rx.Wrap(cli, rx.CombatList{})
	case "CombatJoin":
		return rx.Wrap(cli, rx.CombatJoin{rec["uuid"].(string)})
	case "CombatLeave":
		if !cli.LeaveCombat() {
			log.Warning("Client %s requested to leave combat, but he is not in one.", cli.UUID)
			cli.Send <- tx.Wrap(tx.Error{
				Code:   422,
				Reason: "You cannot leave a combat while not participating one.",
			})
		}
	case "CombatPlayTurn":
		return rx.Wrap(cli, rx.CombatPlayTurn{})
	default:
		// We need to send an error to the client whenever we reach cli point.
		log.Warning("[parser] Client %s has sent a packet containing an invalid code: %s.", cli.UUID, rec)
	}
	return nil
}

// ReadLoop pumps messages from the client to the hub.
func (cli *Client) ReadLoop(cq chan rx.Base) {
	// Whenever the read loop exits, unregister the client from the hub and close the connection.
	defer func() {
		log.Info("Client %s will be disconnected.", cli.UUID)
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

// WriteLoop pumps messages from the hub to the client.
func (cli *Client) WriteLoop() {
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
				log.Warning("Failed to send next packet type to client %s: %s", cli.UUID, err)
				return
			}
			// Try to send the message, and handle failure.
			if _, err := fmt.Fprintf(cli.Conn, "%s\n", json.Encode(message.Command)); err != nil {
				log.Warning("Failed to send packet to client %s: %s", cli.UUID, err)
				return
			}
		case <-ticker.C:
		}
	}
}

func random(min, max int) int {
	return rand.Intn(max-min) + min
}
