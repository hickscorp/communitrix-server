package logic

type Pieces []*Piece

func (this Pieces) Clone() Pieces {
	ret := make(Pieces, len(this))
	for i, piece := range this {
		ret[i] = piece.Clone()
	}
	return ret
}
