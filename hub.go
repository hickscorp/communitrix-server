package main

import (
	"communitrix/cmd/rx"
	"communitrix/cmd/tx"
	"net"
	"reflect"
)

// Hub structure handles interractions between clients.
type Hub struct {
	clients      map[string]*Client // Maintains a list of known clients.
	combats      map[string]*Combat // All existing combats.
	commandQueue chan rx.Base       // Registration, unregistration, subscription, unsubscription, broadcasting.
}

// NewHub is the Hub default constructor.
func NewHub() *Hub {
	hub := &Hub{
		clients:      make(map[string]*Client),
		combats:      make(map[string]*Combat),
		commandQueue: make(chan rx.Base, *config.HubCommandBufferSize),
	}
	// Finally return the fresh hub.
	return hub
}

// RunNewHub is a helper method to easily run a new Hub object.
func RunNewHub() *Hub {
	hub := NewHub()
	go hub.Run()
	return hub
}

// This handler is used from the main program as it's websocket upgrader.
func (hub *Hub) HandleClient(conn net.Conn) {
	// Whenever this method exits, close the connection.
	defer conn.Close()
	// Store the client information for this connection.
	client := hub.NewClient(conn)
	// Send our welcome message.
	client.Send <- tx.Wrap(tx.Welcome{"Hi there!"})
	// Start the writing loop thread, then start reading from the connection.
	go client.WriteLoop()
	client.ReadLoop(hub.commandQueue)
}

// Run is the main loop for any Hub object.
func (hub *Hub) Run() {
	// Loop.
	for {
		// Wait for any event to occur.
		select {
		case command := <-hub.commandQueue:
			client := command.Client.(*Client)
			switch sub := command.Command.(type) {
			// Register a new client.
			case rx.Register:
				log.Info("Client %s registering as %s.", client.UUID, sub.Username)
				client.Username = sub.Username
				client.Send <- tx.Wrap(tx.Registered{sub.Username})

			// Unregisters a client.
			case rx.Unregister:
				log.Info("Client disconnected %s.", client.UUID)
				// Client was in a combat, remove him.
				client.LeaveCombat()
				client.Conn.Close()

			// Client wants a list of existing combats.
			case rx.CombatList:
				combats := make([]string, 0)
				if len(hub.combats) == 0 {
					log.Warning("This server has no combats, creating a default one.")
					combat := hub.RunNewCombat(2, 2)
					hub.combats[combat.UUID] = combat
				}
				for uuid, _ := range hub.combats {
					combats = append(combats, uuid)
				}
				client.Send <- tx.Wrap(tx.CombatList{
					Combats: &combats,
				})
			// Client wants to create a combat.
			case rx.CombatCreate:
				combat := hub.RunNewCombat(sub.MinPlayers, sub.MaxPlayers)
				hub.combats[combat.UUID] = combat
				hub.commandQueue <- *rx.Wrap(client, rx.CombatJoin{UUID: combat.UUID})
			// Client wants to join a combat.
			case rx.CombatJoin:
				combat := hub.combats[sub.UUID]
				if combat == nil {
					log.Warning("The combat %s requested by client %s doesn't exist.", sub.UUID, client.UUID)
					client.Send <- tx.Wrap(tx.Error{
						Code:   404,
						Reason: "Combat was not found.",
					})
				} else {
					client.JoinCombat(combat)
				}
			// Client wants to play his turn in combat.
			case rx.CombatPlayTurn:
				if client.Combat == nil {
					log.Warning("Client %s tried to play a turn in a combat he is not participating to.")
					client.Send <- tx.Wrap(tx.Error{
						Code:   422,
						Reason: "You cannot send turns to combats you're not participating to.",
					})
				}

			// How is that even possible?
			default:
				log.Warning("Client %s sent an unhandled command type: %s.", client.UUID, reflect.TypeOf(sub))
				client.Send <- tx.Wrap(tx.Error{
					Code:   422,
					Reason: "The command you sent could not be understood by the server.",
				})
			}
		}
	}
}
