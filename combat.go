package main

import (
	"fmt"
	"gogs.pierreqr.fr/doodloo/communitrix/cmd/cbt"
	"gogs.pierreqr.fr/doodloo/communitrix/cmd/tx"
	"gogs.pierreqr.fr/doodloo/communitrix/gen"
	"gogs.pierreqr.fr/doodloo/communitrix/i"
	"gogs.pierreqr.fr/doodloo/communitrix/logic"
	"gogs.pierreqr.fr/doodloo/communitrix/util"
	"sync"
)

var (
	combatUUIDMutex       = &sync.Mutex{}
	combatUUID      int64 = 0
)

func NextCombatUUID() int64 {
	combatUUIDMutex.Lock()
	defer combatUUIDMutex.Unlock()
	combatUUID++
	return combatUUID
}

// Player is the base struct representing connected entities.
type Combat struct {
	uuid                   string              // The combat unique identifier on the server.
	mutex                  sync.Mutex          // The lock for this combat.
	players                map[string]i.Player // Maintains a list of known players.
	commandQueue           chan *cbt.Base      // The Combat command queue.
	minPlayers, maxPlayers int                 // The minimum / maximum number of players that can join.
	state                  *combatState        // The current combat state.
}

// This represents the combat state at any point in time.
type combatState struct {
	turn          int                     // The current turn ID.
	target        *logic.Piece            // The objective for all players.
	cells         logic.Pieces            // State of each player
	pieces        logic.Pieces            // The pieces all players are given.
	playedPieces  map[string]map[int]bool // Associative player name -> Piece ID -> Boolean.
	cellsRotation map[string]int          // The current cells belongings.
}

func (this *Combat) UUID() string           { return this.uuid }
func (this *Combat) Notify(cmd interface{}) { this.commandQueue <- cmd.(*cbt.Base) }

func NewCombat(minPlayers, maxPlayers int) *Combat {
	return &Combat{
		uuid:         fmt.Sprintf("CBT%d", NextCombatUUID()),
		mutex:        sync.Mutex{},
		players:      make(map[string]i.Player),
		commandQueue: make(chan *cbt.Base, *config.HubCommandBufferSize),
		minPlayers:   minPlayers,
		maxPlayers:   maxPlayers,
		state:        nil,
	}
}

func (this *Combat) AsSendable() util.MapHelper {
	return util.MapHelper{
		"uuid":       this.uuid,
		"minPlayers": this.minPlayers,
		"maxPlayers": this.maxPlayers,
		"players":    this.sendablePlayers(),
	}
}
func (this *Combat) sendablePlayers() []util.MapHelper {
	players := make([]util.MapHelper, len(this.players))
	idx := 0
	for _, player := range this.players {
		players[idx] = player.AsSendable()
		idx++
	}
	return players
}
func (this *Combat) notifyPlayers(cmd *tx.Base) {
	for _, player := range this.players {
		player.Notify(cmd)
	}
}

func (this *Combat) WhileLocked(do func()) {
	this.mutex.Lock()
	do()
	this.mutex.Unlock()
}

func (this *Combat) Run() {
	// Loop.
	for {
		// Wait for any event to occur.
		select {
		case cmd := <-this.commandQueue:
			switch sub := cmd.Command.(type) {

			// Register a new player.
			case cbt.AddPlayer:
				player := sub.Player.(i.Player)
				if this.state != nil && this.state.turn > 0 {
					player.Notify(tx.Wrap(tx.Error{
						Code:   422,
						Reason: "This combat has already started, you cannot join it anymore.",
					}))
					continue
				}
				if _, ok := this.players[player.UUID()]; !ok {
					// Notify all other players.
					this.notifyPlayers(
						tx.Wrap(tx.CombatPlayerJoined{
							Player: player.AsSendable(),
						}))
					// Add the originator to our list of players.
					this.players[player.UUID()] = player
					// The originator can join.
					player.Notify(tx.Wrap(
						tx.CombatJoin{
							Combat: this.AsSendable(),
						}))
				}
				// We reached the correct number of players, start the combat!
				pCount := len(this.players)
				if pCount == this.maxPlayers { // It's time to start the combat!
					this.commandQueue <- cbt.Wrap(cbt.Prepare{})
				}

			// Unregister a player.
			case cbt.RemovePlayer:
				player := sub.Player.(i.Player)
				delete(this.players, player.UUID())
				// No one left?
				if len(this.players) == 0 {
					log.Warning("There is no one left in combat %s, exiting.", this.uuid)
					return
				}
				// Notify all other players.
				this.notifyPlayers(tx.Wrap(tx.CombatPlayerLeft{UUID: player.UUID()}))

			// Should prepare the combat now.
			case cbt.Prepare:
				if this.state == nil {
					this.state = &combatState{
						turn:          0,
						target:        nil,
						cells:         nil,
						pieces:        nil,
						playedPieces:  make(map[string]map[int]bool),
						cellsRotation: make(map[string]int),
					}

					go func(combat *Combat) {
						notification, ok := combat.Prepare()
						if !ok {
							combat.notifyPlayers(
								tx.Wrap(
									tx.Error{
										Code:   500,
										Reason: "Something went wrong while preparing the combat. Please try again.",
									}))
						} else {
							combat.commandQueue <- cbt.Wrap(*notification)
						}
					}(this)
				}
			// Once the combat is ready... Start it.
			case cbt.Start:
				this.state.turn = 1
				this.state.target, this.state.pieces, this.state.cells = sub.Target, sub.Pieces, sub.Cells
				index := 0
				for _, player := range this.players {
					this.state.cellsRotation[player.UUID()] = index
					this.state.playedPieces[player.UUID()] = make(map[int]bool)
					index++
				}
				this.notifyPlayers(
					tx.Wrap(tx.CombatStart{
						UUID:   this.uuid,
						Target: this.state.target,
						Pieces: this.state.pieces,
						Cells:  this.state.cells,
					}))

			// A new turn has started.
			case cbt.StartNewTurn:

			// A player is playing his turn.
			case cbt.PlayTurn:
				player := sub.Player.(i.Player)
				if this.state == nil || this.state.turn == 0 {
					log.Warning("Client %s is sending turns while the combat hasn't started.", player.UUID())
					player.Notify(tx.Wrap(tx.Error{
						Code:   422,
						Reason: "You cannot play a turn while the combat has not started.",
					}))
					continue
				}
				playedPieces := this.state.playedPieces[player.UUID()]
				if playedPieces[sub.PieceIndex] == true {
					log.Warning("Client %s is trying to play a piece he already played.", player.UUID())
					player.Notify(tx.Wrap(tx.Error{
						Code:   422,
						Reason: "You cannot play the same piece twice.",
					}))
					continue
				}
				// TODO: Check for collisions here.
				playedPieces[sub.PieceIndex] = true
				this.notifyPlayers(
					tx.Wrap(tx.CombatPlayerTurn{
						PlayerUUID: player.UUID(),
						Piece:      this.state.target.Clone().Rotate(sub.Rotation),
					}))
				// Check whether all players have played the current turn.
				allPlayed := true
				for _, playedPieces := range this.state.playedPieces {
					if len(playedPieces) < this.state.turn {
						allPlayed = false
						break
					}
				}
				if allPlayed {
					log.Debug("All players have played their turn. Moving on...")
					this.state.turn++
					this.notifyPlayers(
						tx.Wrap(tx.CombatNewTurn{
							TurnID: this.state.turn,
						}))
				}
			}
		}
	}
}

func (this *Combat) Prepare() (*cbt.Start, bool) {
	log.Warning("Preparing combat %s.", this.uuid)

	// Cache player count.
	playerCount := len(this.players)
	// Prepare data.
	target, ok := gen.NewCellularAutomata(&logic.Vector{5, 7, 5}).Run(0.6)
	if !ok {
		log.Warning("Something went wrong during target generation.")
		return nil, false
	}
	log.Debug("  - Target: Cells %d, Size: %d", target.Size, len(target.Content))

	pieces, ok := gen.NewRecursivePieceSplitter().Run(target, len(target.Content)/15)
	if !ok {
		log.Warning("Something went wrong during pieces generation.")
		return nil, false
	}
	cells, ok := make(logic.Pieces, playerCount), true
	if !ok {
		log.Warning("Something went wrong during cells generation.")
		return nil, false
	}
	// Temporary fix.
	for i := 0; i < playerCount; i++ {
		cells[i] = logic.NewPiece(logic.NewVectorFromInts(0, 0, 0), 0)
	}

	target.CleanUp()
	// Signal combat preparation is over.
	return &cbt.Start{
		Target: target,
		Pieces: pieces,
		Cells:  cells,
	}, true
}
