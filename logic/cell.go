package logic

import (
	"gogs.pierreqr.fr/doodloo/communitrix/util"
)

// Cell represents a point with a value inside a integer coordinate system space.
type Cell struct {
	Vector
	Value int `json:"value"`
}

func NewCellFromInts(x, y, z, v int) *Cell {
	return (&Cell{}).FromInts(x, y, z, v)
}

func (this *Cell) FromInts(x, y, z, v int) *Cell {
	this.X, this.Y, this.Z, this.Value = x, y, z, v
	return this
}

// NewCellFromMap instanciates a new object given a map.
func NewCellFromMap(data util.MapHelper) *Cell {
	return (&Cell{}).FromMap(data)
}

// FromMap replaces the contents of the current object's values by the ones in the given map. The current object is then returned for chaining.
func (this *Cell) FromMap(m util.MapHelper) *Cell {
	this.X, this.Y, this.Z, this.Value = m.Int("x"), m.Int("y"), m.Int("z"), m.Int("value")
	return this
}

// Allows to deep-copy a vector.
func (this *Cell) Clone() *Cell {
	return NewCellFromInts(this.X, this.Y, this.Z, this.Value)
}
