package rx

func Wrap(player interface{}, sub interface{}) *Base {
	return &Base{
		Player:  player,
		Command: sub,
	}
}

type Base struct {
	Player  interface{} // Pointer to a Player.
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
type CombatPlayTurn struct{}
