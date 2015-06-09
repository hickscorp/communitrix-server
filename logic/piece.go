package logic

type Piece struct {
	Size    *Vector   `json:"size"`
	Content []*Vector `json:"content"`
}

func NewEmptyPiece(size *Vector) *Piece {
	return &Piece{
		Size:    size,
		Content: make([]*Vector, 0, size.Volume()),
	}
}

// BreakIntoPieces takes a piece and breaks it down to smaller pieces.
func (this Piece) BreakIntoPieces(fuzyness int) []Piece {
	ret := make([]Piece, 0)
	return ret
}

// Allow to deep-copy a piece.
func (this Piece) Copy() Piece {
	ret := Piece{Size: this.Size}
	ret.Content = make([]*Vector, len(this.Content))
	for i, v := range this.Content {
		ret.Content[i] = v.Copy()
	}
	return ret
}

// Each allows to perform a given function over each of this object's components. The current object is then returned for chaining.
func (this Piece) Each(do func(*Vector)) Piece {
	for _, v := range this.Content {
		do(v)
	}
	return this
}

// Translate applies a given translation transformation to the current object. The current object is then returned for chaining.
func (this Piece) Translate(t *Vector) Piece {
	return this.Each(func(v *Vector) {
		v.Translate(t)
	})
}

// Rotate applies a given rotation transformation to the current object. The current object is then returned for chaining.
func (this Piece) Rotate(q *Quaternion) Piece {
	return this.Each(func(v *Vector) {
		v.Rotate(q)
	})
}
