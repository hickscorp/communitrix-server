package array

import (
	"gogs.pierreqr.fr/doodloo/communitrix/logic"
	"math/rand"
)

// Define our array filling template method.
type ContentArrayFiller func(at *logic.Vector) int

// NewIntContentArrayFiller creates a new routine which will fill an array with a constant value.
func NewIntContentArrayFiller(val int) ContentArrayFiller {
	return func(*logic.Vector) int { return val }
}

// NewRandomContentArrayFiller creates a new routine which will randomly fill an array with values between given bounds.
func NewRandomContentArrayFiller(min int, max int) ContentArrayFiller {
	rnd := max - min + 1
	return func(*logic.Vector) int { return rand.Intn(rnd) + min }
}

// NewCopyContentArrayFiller will copy values from a given tri-ensional array.
func NewCopyContentArrayFiller(other ContentArray) ContentArrayFiller {
	return func(at *logic.Vector) int { return other.Content[at.X][at.Y][at.Z] }
}
