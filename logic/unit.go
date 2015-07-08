package logic

type Unit struct {
	*Piece
	Moves map[string][]int `json:"moves"` // Contains a map from player ID to pieces ID played by this player.
}

type Units []*Unit

func NewEmptyUnit() *Unit {
	return &Unit{
		Piece: NewPiece(NewVectorFromValues(0, 0, 0), 0),
		Moves: make(map[string][]int),
	}
}
