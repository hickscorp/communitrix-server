package main

import (
	"communitrix/cmd/rx"
	"communitrix/cmd/tx"
	"communitrix/i"
	"net"
	"reflect"
)

// Hub structure handles interractions between players.
type Hub struct {
	players      map[string]i.Player // Maintains a list of known players.
	combats      map[string]i.Combat // All existing combats.
	commandQueue chan interface{}    // Registration, unregistration, subscription, unsubscription, broadcasting.
}

// NewHub is the Hub default constructor.
func NewHub() *Hub {
	hub := &Hub{
		players:      make(map[string]i.Player),
		combats:      make(map[string]i.Combat),
		commandQueue: make(chan interface{}, *config.HubCommandBufferSize),
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
	// Store the player information for this connection.
	player := NewPlayer(conn)
	// Send our welcome message.
	player.CommandQueue() <- tx.Wrap(tx.Welcome{Message: "Hi there!"})
	// Start the writing loop thread, then start reading from the connection.
	go player.WriteLoop()
	player.ReadLoop(hub.commandQueue)
}

// Run is the main loop for any Hub object.
func (hub *Hub) Run() {
	// Loop.
	for {
		// Wait for any event to occur.
		select {
		case iCmd := <-hub.commandQueue:
			cmd := iCmd.(rx.Base)
			player := cmd.Player

			switch sub := cmd.Command.(type) {
			// Register a new player.
			case rx.Register:
				log.Info("Player %s registering as %s.", player.UUID(), sub.Username)
				player.SetUsername(sub.Username)
				player.CommandQueue() <- tx.Wrap(tx.Registered{sub.Username})

			// Unregisters a player.
			case rx.Unregister:
				log.Info("Player disconnected %s.", player.UUID())
				// Player was in a combat, remove him.
				player.LeaveCombat()
				player.Connection().Close()

			// Player wants a list of existing combats.
			case rx.CombatList:
				combats := make([]string, 0)
				if len(hub.combats) == 0 {
					log.Warning("This server has no combats, creating a default one.")
					combat := hub.RunNewCombat(2, 6)
					hub.combats[combat.UUID()] = combat
				}
				for uuid, _ := range hub.combats {
					combats = append(combats, uuid)
				}
				player.CommandQueue() <- tx.Wrap(tx.CombatList{
					Combats: &combats,
				})
			// Player wants to create a combat.
			case rx.CombatCreate:
				combat := hub.RunNewCombat(sub.MinPlayers, sub.MaxPlayers)
				hub.combats[combat.UUID()] = combat
				hub.commandQueue <- *rx.Wrap(player, rx.CombatJoin{UUID: combat.UUID()})
			// Player wants to join a combat.
			case rx.CombatJoin:
				combat := hub.combats[sub.UUID]
				if combat == nil {
					log.Warning("The combat %s requested by player %s doesn't exist.", sub.UUID, player.UUID())
					player.CommandQueue() <- tx.Wrap(tx.Error{
						Code:   404,
						Reason: "Combat was not found.",
					})
				} else {
					player.JoinCombat(combat)
				}
			// How is that even possible?
			default:
				log.Warning("Player %s sent an unhandled command type: %s.", player.UUID(), reflect.TypeOf(sub))
				player.CommandQueue() <- tx.Wrap(tx.Error{
					Code:   422,
					Reason: "The command you sent could not be understood by the server.",
				})
			}
		}
	}
}
