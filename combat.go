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
	mutex        sync.Mutex          // The lock for this combat.
	uuid         string              // The combat unique identifier on the server.
	preparing    bool                // Wether this combat is in generation phase.
	started      bool                // Wether this combat has started or not.
	minPlayers   int                 // The minimum number of players that can join.
	maxPlayers   int                 // The maximum number of players that can join.
	players      map[string]i.Player // Maintains a list of known players.
	commandQueue chan *cbt.Base      // The Combat command queue.
	target       *logic.Piece        // The objective for all players.
	cells        []*logic.Piece      // State of each player
	pieces       []*logic.Piece      // The pieces all players are given.
}

func (this *Combat) UUID() string           { return this.uuid }
func (this *Combat) Notify(cmd interface{}) { this.commandQueue <- cmd.(*cbt.Base) }

func NewCombat(minPlayers int, maxPlayers int) *Combat {
	return &Combat{
		mutex:        sync.Mutex{},
		uuid:         fmt.Sprintf("CBT%d", NextCombatUUID()),
		preparing:    false,
		started:      false,
		minPlayers:   minPlayers,
		maxPlayers:   maxPlayers,
		players:      make(map[string]i.Player),
		commandQueue: make(chan *cbt.Base, *config.HubCommandBufferSize),
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
				if this.preparing || this.started {
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
				if !this.preparing && !this.started {
					this.preparing = true
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
				this.preparing, this.started = false, true
				this.target, this.pieces, this.cells = sub.Target, sub.Pieces, sub.Cells
				this.notifyPlayers(
					tx.Wrap(tx.CombatStart{
						UUID:   this.uuid,
						Target: this.target,
						Pieces: this.pieces,
						Cells:  this.cells,
					}))

			// A new turn has started.
			case cbt.StartNewTurn:

			// A player is playing his turn.
			case cbt.PlayTurn:
				player := sub.Player.(i.Player)
				if !this.started {
					log.Warning("Client %s is sending turns while the combat hasn't started.", player.UUID())
					player.Notify(tx.Wrap(tx.Error{
						Code:   422,
						Reason: "You cannot play a turn while the combat has not started.",
					}))
					continue
				}
				piece := this.target.Clone()
				piece.Rotate(sub.Rotation)
				this.notifyPlayers(
					tx.Wrap(tx.CombatPlayerTurn{
						PlayerUUID: player.UUID(),
						Piece:      piece,
					}))
			}
		}
	}
}

func (this *Combat) Prepare() (*cbt.Start, bool) {
	// Cache player count.
	pc := len(this.players)
	// Prepare data.
	target, ok := gen.NewCellularAutomata(&logic.Vector{11, 11, 11}).Run(0.6)
	if !ok {
		log.Warning("Something went wrong during target generation.")
		return nil, false
	}
	log.Info("Finished generating target.")
	pieces, ok := make([]*logic.Piece, pc), true
	if !ok {
		log.Warning("Something went wrong during pieces generation.")
		return nil, false
	}
	cells, ok := make([]*logic.Piece, pc), true
	if !ok {
		log.Warning("Something went wrong during cells generation.")
		return nil, false
	}
	// Temporary fix.
	for i := 0; i < pc; i++ {
		pieces[i] = logic.NewPiece(&logic.Vector{0, 0, 0}, 0)
		cells[i] = logic.NewPiece(&logic.Vector{0, 0, 0}, 0)
	}
	// Signal combat preparation is over.
	return &cbt.Start{
		Target: target,
		Pieces: pieces,
		Cells:  cells,
	}, true
}
