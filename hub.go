package main

import (
	"github.com/hickscorp/communitrix-server/cmd/rx"
	"github.com/hickscorp/communitrix-server/cmd/tx"
	"github.com/hickscorp/communitrix-server/i"
	"github.com/hickscorp/communitrix-server/util"
	"net"
	"reflect"
	"time"
)

// Hub structure handles interractions between players.
type Hub struct {
	players      map[string]i.Player // Maintains a list of known players.
	combats      map[string]i.Combat // All existing combats.
	commandQueue chan *rx.Base       // Registration, unregistration, subscription, unsubscription, broadcasting.
}

// NewHub is the Hub default constructor.
func NewHub() *Hub {
	return &Hub{
		players:      make(map[string]i.Player),
		combats:      make(map[string]i.Combat),
		commandQueue: make(chan *rx.Base, *config.HubCommandBufferSize),
	}
}

// This handler is used from the main program as it's websocket upgrader.
func (this *Hub) HandleClient(conn net.Conn) {
	log.Debug("New client connected, spawning routine.")
	time.Sleep(time.Second * 1)
	// Whenever this method exits, close the connection.
	defer conn.Close()
	// Store the player information for this connection.
	StartNewPlayer(this.commandQueue, conn)
}

// Run is the main loop for any Hub object.
func (this *Hub) Run() {
	log.Debug("Running new hub.")
	// Loop.
	for {
		// Wait for any event to occur.
		select {
		case cmd := <-this.commandQueue:
			player := cmd.Player

			switch sub := cmd.Command.(type) {
			// Register a new player.
			case rx.Register:
				log.Debug("Player %s registering as %s.", player.UUID(), sub.Username)
				player.SetUsername(sub.Username)
				player.Notify(tx.Wrap(tx.Registered{sub.Username}))

			// Unregisters a player.
			case rx.Unregister:
				log.Debug("Player disconnected %s.", player.UUID())
				// Player was in a combat, remove him.
				player.LeaveCombat()
				player.Connection().Close()

			// Player wants a list of existing combats.
			case rx.CombatList:
				time.Sleep(time.Second * 1)
				// TODO: Remove this from there!!!
				for len(this.combats) < 2 {
					log.Warning("This server only has %d combats, creating one more.", len(this.combats))
					combat := NewCombat(1, 1)
					this.combats[combat.UUID()] = combat
					go func(combat *Combat, ch chan<- *rx.Base) {
						combat.Run()
						ch <- rx.Wrap(nil, rx.CombatEnd{UUID: combat.uuid})
					}(combat, this.commandQueue)
				}
				// Retrieve a list of combats.
				combats := make([]util.MapHelper, 0)
				comms := make(chan util.MapHelper)
				for _, combat := range this.combats {
					go combat.Summarize(comms)
					summary := <-comms
					if !summary["started"].(bool) {
						combats = append(combats, summary)
					}
				}
				player.Notify(tx.Wrap(tx.CombatList{Combats: combats}))

			// Player wants to create a combat.
			case rx.CombatCreate:
				combat := NewCombat(sub.MinPlayers, sub.MaxPlayers)
				this.combats[combat.UUID()] = combat
				go func(combat *Combat, ch chan<- *rx.Base) {
					combat.Run()
					ch <- rx.Wrap(nil, rx.CombatEnd{UUID: combat.uuid})
				}(combat, this.commandQueue)
				this.commandQueue <- rx.Wrap(player, rx.CombatJoin{UUID: combat.UUID()})

			// Player wants to join a combat.
			case rx.CombatJoin:
				combat := this.combats[sub.UUID]
				if combat == nil {
					log.Warning("The combat %s requested by player %s doesn't exist.", sub.UUID, player.UUID())
					player.Notify(tx.Wrap(tx.Error{
						Code:   404,
						Reason: "Combat was not found.",
					}))
					continue
				}
				player.JoinCombat(combat)

			// A combat has ended.
			case rx.CombatEnd:
				if _, ok := this.combats[sub.UUID]; ok {
					delete(this.combats, sub.UUID)
				}

			// How is that even possible?
			default:
				log.Warning("Player %s sent an unhandled command type: %s.", player.UUID(), reflect.TypeOf(sub))
				player.Notify(tx.Wrap(tx.Error{
					Code:   422,
					Reason: "The command you sent could not be understood by the server.",
				}))
			}
		}
	}
}
