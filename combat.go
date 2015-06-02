package main

import (
	"communitrix/cmd/cbt"
	"communitrix/cmd/tx"
	"fmt"
	"sync"
)

var combatUUIDMutex = &sync.Mutex{}
var combatUUID int64 = 0

func NextCombatUUID() int64 {
	combatUUIDMutex.Lock()
	defer combatUUIDMutex.Unlock()
	combatUUID++
	return combatUUID
}

// Client is the base struct representing connected entities.
type Combat struct {
	UUID         string           // The combat unique identifier on the server.
	Started      bool             // Wether this combat has started or not.
	MinPlayers   int              // The minimum number of players that can join.
	MaxPlayers   int              // The maximum number of players that can join.
	clients      map[*Client]bool // Maintains a list of known clients.
	CommandQueue chan *cbt.Base   // The Combat command queue.
}

// NewClient is exposed on the Hub class.
func (hub *Hub) RunNewCombat(minPlayers int, maxPlayers int) *Combat {
	ret := &Combat{
		UUID:         fmt.Sprintf("CBT%d", NextCombatUUID()),
		Started:      false,
		MinPlayers:   minPlayers,
		MaxPlayers:   maxPlayers,
		clients:      make(map[*Client]bool),
		CommandQueue: make(chan *cbt.Base, *config.HubCommandBufferSize),
	}
	go ret.Run()
	return ret
}

func (combat *Combat) Run() {
	log.Info("Starting combat %s loop.", combat.UUID)
	// Loop.
	for {
		// Wait for any event to occur.
		select {
		case cmd := <-combat.CommandQueue:
			switch sub := cmd.Command.(type) {

			// Register a new client.
			case cbt.AddClient:
				client := sub.Client.(*Client)
				if _, ok := combat.clients[client]; !ok {
					// Notify all other clients.
					notification := tx.Wrap(tx.CombatPlayerJoined{Player: client.UUID})
					for otherClient, _ := range combat.clients {
						otherClient.Send <- notification
					}
					// Add the originator to our list of clients.
					combat.clients[client] = true
					// The originator can join.
					client.Send <- tx.Wrap(tx.CombatJoin{
						UUID:       combat.UUID,
						MinPlayers: combat.MinPlayers,
						MaxPlayers: combat.MaxPlayers,
						Players:    combat.PlayerList(),
					})
				}
				// We reached the correct number of players, start the combat!
				if combat.MaxPlayers <= len(combat.clients) {
					notification := tx.Wrap(tx.CombatStart{
						UUID:    combat.UUID,
						Players: combat.PlayerList(),
					})
					for otherClient, _ := range combat.clients {
						otherClient.Send <- notification
					}
				}

			// Unregister a client.
			case cbt.RemoveClient:
				client := sub.Client.(*Client)
				delete(combat.clients, sub.Client.(*Client))
				// Notify all other clients.
				notification := tx.Wrap(tx.CombatPlayerLeft{Player: client.UUID})
				for otherClient, _ := range combat.clients {
					otherClient.Send <- notification
				}
			}
		}
	}
}

// Builds a players list.
func (combat *Combat) PlayerList() *[]string {
	ret := make([]string, 0)
	for client, _ := range combat.clients {
		ret = append(ret, client.UUID)
	}
	return &ret
}
