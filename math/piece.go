package math

import (
	"github.com/op/go-logging"
	"math/rand"
)

var log = logging.MustGetLogger("communitrix")

type Piece []*Vector

func NewSamplePiece() *Piece {
	return &Piece{
		&Vector{0, 0, 0},
		&Vector{1, 0, 0},
		&Vector{2, 0, 0},
		&Vector{1, 1, 0},
	}
}

func NewRandomPiece(size *Vector, density int32) *Piece {
	ret := make(Piece, 0)
	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			for z := 0; z < size.Z; z++ {
				rnd := rand.Int31n(100)
				if rnd <= density {
					ret = append(ret, &Vector{x, y, z})
				}
			}
		}
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
