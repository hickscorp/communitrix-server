package main

import (
	"communitrix/cmd/cbt"
	"communitrix/cmd/tx"
	"communitrix/util"
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
	mutex        sync.Mutex       // The lock for this combat.
	uuid         string           // The combat unique identifier on the server.
	started      bool             // Wether this combat has started or not.
	minPlayers   int              // The minimum number of players that can join.
	maxPlayers   int              // The maximum number of players that can join.
	players      map[*Player]bool // Maintains a list of known players.
	commandQueue chan interface{} // The Combat command queue.
}

func (combat *Combat) UUID() string                   { return combat.uuid }
func (combat *Combat) CommandQueue() chan interface{} { return combat.commandQueue }

func NewCombat(minPlayers int, maxPlayers int) *Combat {
	return &Combat{
		mutex:        sync.Mutex{},
		uuid:         fmt.Sprintf("CBT%d", NextCombatUUID()),
		started:      false,
		minPlayers:   minPlayers,
		maxPlayers:   maxPlayers,
		players:      make(map[*Player]bool),
		commandQueue: make(chan interface{}, *config.HubCommandBufferSize),
	}
}
func (hub *Hub) RunNewCombat(minPlayers int, maxPlayers int) *Combat {
	ret := NewCombat(minPlayers, maxPlayers)
	go ret.Run()
	return ret
}

func (combat *Combat) AsSendable() *util.JsonMap {
	return &util.JsonMap{
		"uuid":       combat.uuid,
		"minPlayers": combat.minPlayers,
		"maxPlayers": combat.maxPlayers,
		"players":    combat.SendablePlayers(),
	}
}
func (combat *Combat) SendablePlayers() *[]*util.JsonMap {
	players := make([]*util.JsonMap, len(combat.players))
	idx := 0
	for player, _ := range combat.players {
		players[idx] = player.AsSendable()
		idx++
	}
	return &players
}

func (combat *Combat) WhileLocked(do func()) {
	combat.mutex.Lock()
	do()
	combat.mutex.Unlock()
}

func (combat *Combat) Run() {
	log.Info("Starting combat %s loop.", combat.uuid)
	// Loop.
	for {
		// Wait for any event to occur.
		select {
		case iCmd := <-combat.commandQueue:
			switch sub := iCmd.(*cbt.Base).Command.(type) {

			// Register a new player.
			case cbt.AddPlayer:
				player := sub.Player.(*Player)
				if _, ok := combat.players[player]; !ok {
					// Notify all other players.
					notification := tx.Wrap(tx.CombatPlayerJoined{Player: player.AsSendable()})
					for otherPlayer, _ := range combat.players {
						otherPlayer.CommandQueue() <- notification
					}
					// Add the originator to our list of players.
					combat.players[player] = true
					// The originator can join.
					player.CommandQueue() <- tx.Wrap(tx.CombatJoin{Combat: combat.AsSendable()})
				}
				// We reached the correct number of players, start the combat!
				if combat.maxPlayers <= len(combat.players) {
					notification := tx.Wrap(tx.CombatStart{
						UUID:    combat.uuid,
						Players: combat.PlayerList(),
					})
					for otherPlayer, _ := range combat.players {
						otherPlayer.CommandQueue() <- notification
					}
				}

			// Unregister a player.
			case cbt.RemovePlayer:
				player := sub.Player.(*Player)
				delete(combat.players, sub.Player.(*Player))
				// Notify all other players.
				notification := tx.Wrap(tx.CombatPlayerLeft{UUID: player.UUID()})
				for otherPlayer, _ := range combat.players {
					otherPlayer.CommandQueue() <- notification
				}
			}
		}
	}
}

// Builds a players list.
func (combat *Combat) PlayerList() *[]string {
	ret := make([]string, 0)
	for player, _ := range combat.players {
		ret = append(ret, player.UUID())
	}
	return &ret
}
