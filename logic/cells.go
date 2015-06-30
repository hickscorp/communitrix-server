package logic

import "math/rand"

type Cells []*Cell

func (this Cells) Clone() Cells {
	ret := make(Cells, len(this))
	for i, cell := range this {
		ret[i] = cell.Clone()
	}
	return ret
}

func (this Cells) Shuffle() Cells {
  for i := len(this) - 1; i > 0; i-- {
    j := rand.Intn(i)
    this[i], this[j] = this[j], this[i]
  }
  return this
}

func (this Cells) CollidesWith(c *Cell) bool {
  for _, cell := range this {
    if cell.Vector.CollidesWith(c.Vector) {
      return true
    }
  }
  return false
}
