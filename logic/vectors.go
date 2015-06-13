package logic

import "math/rand"

type Vectors []*Vector

func (this Vectors) Clone() Vectors {
	ret := make(Vectors, len(this))
	for i, cell := range this {
		ret[i] = cell.Clone()
	}
	return ret
}

func (this Vectors) Shuffle() Vectors {
	for i := len(this) - 1; i > 0; i-- {
		j := rand.Intn(i)
		this[i], this[j] = this[j], this[i]
	}
	return this
}
