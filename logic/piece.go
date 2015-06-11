package logic

type Piece struct {
	Size    *Vector `json:"size"`
	Content []*Cell `json:"content"`
}

func NewPiece(size *Vector, capacity int) *Piece {
	return &Piece{
		Size:    size,
		Content: make([]*Cell, 0, capacity),
	}
}

// Copy allows to deep-copy a piece.
func (this Piece) Clone() *Piece {
	ret := Piece{Size: this.Size.Clone()}
	ret.Content = make([]*Cell, len(this.Content))
	for i, v := range this.Content {
		ret.Content[i] = v.Clone()
	}
	return &ret
}

// Translate applies a given translation transformation to the current object. The current object is then returned for chaining.
func (this *Piece) Translate(t *Vector) {
	for _, v := range this.Content {
		v.Translate(t)
	}
}

// Rotate applies a given rotation transformation to the current object. The current object is then returned for chaining.
func (this *Piece) Rotate(q *Quaternion) {
	this.Size.Rotate(q)
	for _, v := range this.Content {
		v.Rotate(q)
	}
}

func (this *Piece) Each(do func(*Cell)) {
	for _, v := range this.Content {
		do(v)
	}
}
