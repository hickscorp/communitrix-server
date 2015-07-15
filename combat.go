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
	players                map[string]i.Player // Maintains a list of known players.
	commandQueue           chan *cbt.Base      // The Combat command queue.
	minPlayers, maxPlayers int                 // The minimum / maximum number of players that can join.
	state                  *combatState        // The current combat state.
}

// This represents the combat state at any point in time.
type combatState struct {
	turn          int                     // The current turn ID.
	target        *logic.Piece            // The objective for all players.
	units         logic.Units             // State of each unit.
	pieces        logic.Pieces            // The pieces all players are given.
	playedPieces  map[string]map[int]bool // Associative player name -> Piece ID -> Boolean.
	playerIndices map[string]int          // Indices associated to each player.
}

func (this *Combat) UUID() string           { return this.uuid }
func (this *Combat) Notify(cmd interface{}) { this.commandQueue <- cmd.(*cbt.Base) }

func NewCombat(minPlayers, maxPlayers int) *Combat {
	return &Combat{
		uuid:         fmt.Sprintf("CBT%d", NextCombatUUID()),
		players:      make(map[string]i.Player),
		commandQueue: make(chan *cbt.Base, *config.HubCommandBufferSize),
		minPlayers:   minPlayers,
		maxPlayers:   maxPlayers,
		state:        nil,
	}
}

func (this *Combat) Summarize(ret chan util.MapHelper) {
	this.commandQueue <- cbt.Wrap(cbt.Summarize{Ret: ret})
}
func (this *Combat) AsSendable() util.MapHelper {
	turn := 0
	if this.state != nil {
		turn = this.state.turn
	}
	return util.MapHelper{
		"uuid":        this.uuid,
		"minPlayers":  this.minPlayers,
		"maxPlayers":  this.maxPlayers,
		"started":     this.state != nil,
		"currentTurn": turn,
		"players":     this.sendablePlayers(),
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
func (this *Combat) notifyPlayers(do func(i.Player) *tx.Base, perPlayerNotification bool) {
	if perPlayerNotification {
		for _, player := range this.players {
			player.Notify(do(player))
		}
	} else {
		notif := do(nil)
		for _, player := range this.players {
			player.Notify(notif)
		}
	}
}

func (this *Combat) Run() {
	// Loop.
	for {
		// Wait for any event to occur.
		select {
		case cmd := <-this.commandQueue:
			switch sub := cmd.Command.(type) {

			// Get a summary about this combat.
			case cbt.Summarize:
				sub.Ret <- this.AsSendable()

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
					joinNotif := func(i.Player) *tx.Base {
						return tx.Wrap(tx.CombatPlayerJoined{
							Player: player.AsSendable(),
						})
					}
					this.notifyPlayers(joinNotif, false)
					// Add the originator to our list of players.
					this.players[player.UUID()] = player
					// The originator can join.
					player.Notify(tx.Wrap(tx.CombatJoin{Combat: this.AsSendable()}))
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
				leaveNotif := func(i.Player) *tx.Base {
					return tx.Wrap(tx.CombatPlayerLeft{UUID: player.UUID()})
				}
				this.notifyPlayers(leaveNotif, false)

			// Should prepare the combat now.
			case cbt.Prepare:
				if this.state == nil {
					this.state = &combatState{
						turn:          0,
						target:        nil,
						units:         nil,
						pieces:        nil,
						playedPieces:  make(map[string]map[int]bool),
						playerIndices: make(map[string]int),
					}
					// Create players indices mapping.
					idx := 0
					for uuid, _ := range this.players {
						this.state.playerIndices[uuid] = idx
						idx++
					}

					notification, ok := this.Prepare()
					if !ok {
						errorNotif := func(i.Player) *tx.Base {
							return tx.Wrap(tx.Error{Code: 500, Reason: "Something went wrong while preparing the combat. Please try again."})
						}
						this.notifyPlayers(errorNotif, false)
					} else {
						this.commandQueue <- cbt.Wrap(*notification)
					}
				}
			// Once the combat is ready... Start it.
			case cbt.Start:
				this.state.turn = 1
				this.state.target, this.state.pieces, this.state.units = sub.Target, sub.Pieces, sub.Units
				index := 0
				for _, player := range this.players {
					this.state.playedPieces[player.UUID()] = make(map[int]bool)
					index++
				}
				// Give everyone the combat start notification.
				startNotif := func(i.Player) *tx.Base {
					return tx.Wrap(
						tx.CombatStart{
							UUID:   this.uuid,
							Target: this.state.target,
							Units:  this.state.units,
							Pieces: this.state.pieces,
						})
				}
				this.notifyPlayers(startNotif, false)
				// Give everyone informations about the turn.
				turnNotif := func(p i.Player) *tx.Base {
					return tx.Wrap(tx.CombatNewTurn{
						TurnID: this.state.turn,
						UnitID: (this.state.playerIndices[p.UUID()] + this.state.turn) % len(this.state.units),
					})
				}
				this.notifyPlayers(turnNotif, true)

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

				angles := sub.Rotation.ToEulerAngles()
				if !angles.IsMultipleOf(90) {
					log.Warning("Wrong rotation detected with Quaternion %v: %v.", sub.Rotation, angles)
					player.Notify(tx.Wrap(tx.Acknowledgment{
						Serial:       "PlayTurn",
						Valid:        false,
						ErrorMessage: "An invalid rotation was detected. Please play again.",
					}))
					continue
				}

				piece := this.state.pieces[sub.PieceIndex].Clone().Rotate(sub.Rotation).Translate(sub.Translation)
				log.Debug("Piece %d played with translation %v and rotation %v.", sub.PieceIndex, sub.Translation, sub.Rotation)

				// TODO: Check for collisions here.
				playerIndex := this.state.playerIndices[player.UUID()]
				unitId := (playerIndex + this.state.turn) % len(this.state.units)
				unit := this.state.units[unitId]

				for _, v := range piece.Content {
					if unit.Content.CollidesWith(v) {
						unit = nil
						break
					}
				}
				player.Notify(tx.Wrap(tx.Acknowledgment{
					Serial:       "PlayTurn",
					Valid:        unit != nil,
					ErrorMessage: "A collision was detected. Please play again.",
				}))
				if unit == nil {
					log.Warning("Collision detected")
					break
				}
				// Register the piece as played for this player.
				playedPieces[sub.PieceIndex] = true
				// Merge the played piece into the current unit.
				for _, c := range piece.Content {
					unit.AddCell(c)
				}
				// Clean up the unit.
				unit.CleanUp()
				// Keep track of the played piece ID within this unit.
				ids := unit.Moves[player.UUID()]
				if ids == nil {
					ids = make([]int, 0)
				}
				ids = append(ids, sub.PieceIndex)
				unit.Moves[player.UUID()] = ids

				playerMoveNotif := func(i.Player) *tx.Base {
					return tx.Wrap(tx.CombatPlayerTurn{
						PlayerUUID: player.UUID(),
						PieceID:    sub.PieceIndex,
						UnitID:     unitId,
						Unit:       unit,
					})
				}
				this.notifyPlayers(playerMoveNotif, false)

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
					for _, player := range this.players {
						player.Notify(
							tx.Wrap(tx.CombatNewTurn{
								TurnID: this.state.turn,
								UnitID: (this.state.playerIndices[player.UUID()] + this.state.turn) % len(this.state.units),
							}))
					}
					if this.state.turn > len(this.state.pieces) {
						log.Warning("LAST TURN WAS JUST PLAYED.")
						// Notify all other players.
						endNotif := func(i.Player) *tx.Base {
							return tx.Wrap(tx.CombatEnd{})
						}
						this.notifyPlayers(endNotif, false)
						return
					}
				}
			}
		}
	}
}

func (this *Combat) Prepare() (*cbt.Start, bool) {
	log.Debug("Preparing combat %s.", this.uuid)

	// Cache player count.
	playerCount := len(this.players)
	// Prepare data.
	target, ok := gen.NewCellularAutomata(&logic.Vector{4, 3, 3}).Run(0.5)
	if !ok {
		log.Warning("Something went wrong during target generation.")
		return nil, false
	}
	log.Debug("  - Target: Cells %d, Size: %d", target.Size, len(target.Content))

	pieces, ok := gen.NewPieceSplitter().Run(target, 6)
	if !ok {
		log.Warning("Something went wrong during pieces generation.")
		return nil, false
	}
	units, ok := make(logic.Units, playerCount), true
	if !ok {
		log.Warning("Something went wrong during units generation.")
		return nil, false
	}
	// Temporary fix.
	for i := 0; i < playerCount; i++ {
		units[i] = logic.NewEmptyUnit() //logic.NewPiece(logic.NewVectorFromValues(0, 0, 0), 0)
	}

	target.CleanUp()
	// Signal combat preparation is over.
	return &cbt.Start{
		Target: target,
		Pieces: pieces,
		Units:  units,
	}, true
}
