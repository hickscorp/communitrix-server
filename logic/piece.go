package logic

import "math/rand"

type Piece []*Vector

func NewRandomPiece(size *Vector, density int) *Piece {
	ret := make(Piece, 0)
	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			for z := 0; z < size.Z; z++ {
				rnd := rand.Intn(100)
				if rnd <= density {
				}
			}
		}
	}
	return &ret
}

func (this *Piece) BreakIntoPieces (fuzyness int) *[]*Piece {
	ret := make([]*Piece, 0)
	return &ret;
}

// Allow to deep-copy a piece.
func (this *Piece) Copy() *Piece {
	ret := make(Piece, len(*this))
	for idx, v := range *this {
		ret[idx] = v.Copy()
	}
	return &ret
}

// Each allows to perform a given function over each of this object's components. The current object is then returned for chaining.
func (this *Piece) Each(do func(*Vector)) *Piece {
	for _, v := range *this {
		do(v)
	}
	return this
}

// Translate applies a given translation transformation to the current object. The current object is then returned for chaining.
func (this *Piece) Translate(t *Vector) *Piece {
	return this.Each(func(v *Vector) {
		v.Translate(t)
	})
}

// Rotate applies a given rotation transformation to the current object. The current object is then returned for chaining.
func (this *Piece) Rotate(q *Quaternion) *Piece {
	return this.Each(func(v *Vector) {
		v.Rotate(q)
	})
}
