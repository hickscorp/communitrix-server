package logic

type Piece struct {
	Size    *Vector   `json:"size"`
	Content []*Vector `json:"content"`
}

func NewPiece(size *Vector, capacity int) *Piece {
	return &Piece{
		Size:    size,
		Content: make([]*Vector, 0, capacity),
	}
}

// Copy allows to deep-copy a piece.
func (this Piece) Copy() *Piece {
	ret := Piece{Size: this.Size}
	ret.Content = make([]*Vector, len(this.Content))
	for i, v := range this.Content {
		ret.Content[i] = v.Copy()
	}
	return &ret
}

// Translate applies a given translation transformation to the current object. The current object is then returned for chaining.
func (this *Piece) Translate(t *Vector) *Piece {
	for _, v := range this.Content {
		v.Translate(t)
	}
	return this
}

// Rotate applies a given rotation transformation to the current object. The current object is then returned for chaining.
func (this *Piece) Rotate(q *Quaternion) *Piece {
	for _, v := range this.Content {
		v.Rotate(q)
	}
	return this
}
