package main

import (
	"communitrix/cmd/cbt"
	"communitrix/cmd/tx"
	"communitrix/i"
	"communitrix/math"
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
	mutex        sync.Mutex          // The lock for this combat.
	uuid         string              // The combat unique identifier on the server.
	started      bool                // Wether this combat has started or not.
	minPlayers   int                 // The minimum number of players that can join.
	maxPlayers   int                 // The maximum number of players that can join.
	players      map[string]i.Player // Maintains a list of known players.
	commandQueue chan interface{}    // The Combat command queue.
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
		players:      make(map[string]i.Player),
		commandQueue: make(chan interface{}, *config.HubCommandBufferSize),
	}
}

func (combat *Combat) AsSendable() *util.JsonMap {
	return &util.JsonMap{
		"uuid":       combat.uuid,
		"minPlayers": combat.minPlayers,
		"maxPlayers": combat.maxPlayers,
		"players":    combat.sendablePlayers(),
	}
}
func (combat *Combat) sendablePlayers() *[]*util.JsonMap {
	players := make([]*util.JsonMap, len(combat.players))
	idx := 0
	for _, player := range combat.players {
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
				if combat.started {
					sub.Player.CommandQueue() <- tx.Wrap(tx.Error{
						Code:   422,
						Reason: "This combat has already started, you cannot join it anymore.",
					})
					continue
				}
				if _, ok := combat.players[sub.Player.UUID()]; !ok {
					// Notify all other players.
					notification := tx.Wrap(tx.CombatPlayerJoined{Player: sub.Player.AsSendable()})
					for _, otherPlayer := range combat.players {
						otherPlayer.CommandQueue() <- notification
					}
					// Add the originator to our list of players.
					combat.players[sub.Player.UUID()] = sub.Player
					// The originator can join.
					sub.Player.CommandQueue() <- tx.Wrap(tx.CombatJoin{Combat: combat.AsSendable()})
				}
				// We reached the correct number of players, start the combat!
				pCount := len(combat.players)
				if pCount == combat.maxPlayers { // It's time to start the combat!
					combat.Start()
				} else if pCount > combat.maxPlayers { // Impossible case. Just put some logging to make sure.
					log.Error("BUG: There %d / %d players in combat %s.", pCount, combat.maxPlayers, combat.uuid)
				}

			// Unregister a player.
			case cbt.RemovePlayer:
				if combat.started {
					sub.Player.CommandQueue() <- tx.Wrap(tx.Error{
						Code:   422,
						Reason: "This combat has already started, you cannot join it anymore.",
					})
					continue
				}
				delete(combat.players, sub.Player.UUID())
				// Notify all other players.
				notification := tx.Wrap(tx.CombatPlayerLeft{UUID: sub.Player.UUID()})
				for _, otherPlayer := range combat.players {
					otherPlayer.CommandQueue() <- notification
				}

			case cbt.Start:
				combat.Start()
			}
		}
	}
}

func (combat *Combat) Start() {
	if combat.started {
		log.Error("Combat cannot be started, as it has already started.")
		return
	}
	combat.started = true

	// Generate a random fuel cell.
	target := math.NewRandomPiece(&math.Vector{3, 3, 3}, 50)
	pieces := &[]math.Piece{}
	// Break the fuel cell into pieces.
	// ...

	n1 := tx.Wrap(tx.CombatStart{
		UUID:   combat.uuid,
		Target: &target,
		Pieces: &pieces,
	})
	n2 := tx.Wrap(tx.CombatPlayerTurn{
		Contents: target,
	})
	for _, otherPlayer := range combat.players {
		otherPlayer.CommandQueue() <- n1
		otherPlayer.CommandQueue() <- n2
	}
}
