package rx

import "gogs.pierreqr.fr/doodloo/communitrix/i"

func Wrap(player i.Player, sub interface{}) *Base {
	return &Base{
		Player:  player,
		Command: sub,
	}
}

type Base struct {
	Player  i.Player    // Pointer to a Player.
	Command interface{} // Command.
}

type Register struct {
	Username string
}
type Unregister struct{}

type CombatList struct{}
type CombatCreate struct {
	MinPlayers int
	MaxPlayers int
}
type CombatJoin struct {
	UUID string
}
type CombatLeave struct{}
type CombatPlayTurn struct {
	UUID        string
	Rotation    interface{}
	Translation interface{}
}
type CombatEnd struct {
	UUID string
}
