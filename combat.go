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

// Player is the base struct representing connected entities.
type Combat struct {
	UUID         string           // The combat unique identifier on the server.
	Started      bool             // Wether this combat has started or not.
	MinPlayers   int              // The minimum number of players that can join.
	MaxPlayers   int              // The maximum number of players that can join.
	Players      map[*Player]bool // Maintains a list of known players.
	CommandQueue chan *cbt.Base   // The Combat command queue.
}

func (combat *Combat) AsSendable() *map[string]interface{} {
	return &map[string]interface{}{
		"uuid":       combat.UUID,
		"minPlayers": combat.MinPlayers,
		"maxPlayers": combat.MaxPlayers,
		"players":    combat.SendablePlayers(),
	}
}
func (combat *Combat) SendablePlayers() *[]*map[string]interface{} {
	players := make([]*map[string]interface{}, len(combat.Players))
	idx := 0
	for player, _ := range combat.Players {
		players[idx] = player.AsSendable()
		idx++
	}
	return &players
}

// NewPlayer is exposed on the Hub class.
func (hub *Hub) RunNewCombat(minPlayers int, maxPlayers int) *Combat {
	ret := &Combat{
		UUID:         fmt.Sprintf("CBT%d", NextCombatUUID()),
		Started:      false,
		MinPlayers:   minPlayers,
		MaxPlayers:   maxPlayers,
		Players:      make(map[*Player]bool),
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

			// Register a new player.
			case cbt.AddPlayer:
				player := sub.Player.(*Player)
				if _, ok := combat.Players[player]; !ok {
					// Notify all other players.
					notification := tx.Wrap(tx.CombatPlayerJoined{Player: player.AsSendable()})
					for otherPlayer, _ := range combat.Players {
						otherPlayer.Send <- notification
					}
					// Add the originator to our list of players.
					combat.Players[player] = true
					// The originator can join.
					player.Send <- tx.Wrap(tx.CombatJoin{Combat: combat.AsSendable()})
				}
				// We reached the correct number of players, start the combat!
				if combat.MaxPlayers <= len(combat.Players) {
					notification := tx.Wrap(tx.CombatStart{
						UUID:    combat.UUID,
						Players: combat.PlayerList(),
					})
					for otherPlayer, _ := range combat.Players {
						otherPlayer.Send <- notification
					}
				}

			// Unregister a player.
			case cbt.RemovePlayer:
				player := sub.Player.(*Player)
				delete(combat.Players, sub.Player.(*Player))
				// Notify all other players.
				notification := tx.Wrap(tx.CombatPlayerLeft{UUID: player.UUID})
				for otherPlayer, _ := range combat.Players {
					otherPlayer.Send <- notification
				}
			}
		}
	}
}

// Builds a players list.
func (combat *Combat) PlayerList() *[]string {
	ret := make([]string, 0)
	for player, _ := range combat.Players {
		ret = append(ret, player.UUID)
	}
	return &ret
}
