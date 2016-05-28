package array

import "github.com/hickscorp/communitrix-server/logic"

// Define our array filling template method.
type ContentArrayFiller func(at *logic.Vector) int

// NewIntContentArrayFiller creates a new routine which will fill an array with a constant value.
func NewIntContentArrayFiller(val int) ContentArrayFiller {
	return func(*logic.Vector) int { return val }
}

// NewCopyContentArrayFiller will copy values from a given tri-ensional array.
func NewCopyContentArrayFiller(other *ContentArray) ContentArrayFiller {
	return func(at *logic.Vector) int { return other.Content[at.X][at.Y][at.Z] }
}
