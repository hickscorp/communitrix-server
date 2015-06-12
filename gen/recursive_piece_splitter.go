package gen

import (
	"gogs.pierreqr.fr/doodloo/communitrix/array"
	"gogs.pierreqr.fr/doodloo/communitrix/logic"
	"math/rand"
)

type RecursivePieceSplitter struct {
}

func NewRecursivePieceSplitter() *RecursivePieceSplitter {
	return &RecursivePieceSplitter{}
}

func (this *RecursivePieceSplitter) Run(piece *logic.Piece, count int) (logic.Pieces, bool) {
	// params validation
	if piece == nil {
		log.Warning("PieceSplitter cannot run without a piece.")
		return nil, false
	} else if count <= 0 {
		log.Warning("PieceSplitter break down a piece into a negative count of pieces.")
		return nil, false
	} else if count >= len(piece.Content) {
		log.Warning("PieceSplitter cannot break down a piece to more than its count of cells.")
		return nil, false
	}

	// Convert the piece to an array.
	arr := array.NewContentArrayFromPiece(piece)
	log.Warning("Total number of pieces: %d", len(piece.Content))
	// Prepare starting points.
	ats := make([]*logic.Vector, 0, count)
	for i := 0; i < count; i++ {
		at := logic.NewVectorFromInts(rand.Intn(piece.Size.X), rand.Intn(piece.Size.Y), rand.Intn(piece.Size.Z))
		if at.SameAsAny(ats) {
			i--
			continue
		}
		ats = append(ats, at)
	}

	// Start as many goroutines as we have starting points.
	recursors := make([]*Recursor, count)
	for i, at := range ats {
		recursor := NewRecursor(i+1, arr)
		arr.Content[at.X][at.Y][at.Z] = recursor.ID
		go recursor.Recurse(at, 0)
		recursors[i] = recursor
	}
	done := 0
	for i := 0; done < count; i++ {
		idx := i % count
		recursor := recursors[idx]
		if !recursor.AllDone {
			recursor.Advance <- true
			if !<-recursor.Done {
				recursor.AllDone = true
				done++
			}
		}
	}
	for _, recursor := range recursors {
		log.Warning("Recursor %d has %d piece(s).", recursor.ID, recursor.Pieces)
	}

	ret := make(logic.Pieces, count)
	for i, recursor := range recursors {
		ret[i] = logic.NewPiece(logic.NewVectorFromInts(10, 10, 10), recursor.Pieces)
	}
	arr.Each(func(at *logic.Vector, value int) {
		if value > 0 {
			ret[value-1].Content = append(ret[value-1].Content, logic.NewCellFromInts(at.X, at.Y, at.Z, 1))
		}
	})
	return ret, true
}

type Recursor struct {
	ID      int
	Arr     *array.ContentArray
	Advance chan bool
	Done    chan bool
	Pieces  int
	AllDone bool
}

func NewRecursor(id int, arr *array.ContentArray) *Recursor {
	return &Recursor{
		ID:      id,
		Arr:     arr,
		Advance: make(chan bool, 2),
		Done:    make(chan bool),
		Pieces:  1,
		AllDone: false,
	}
}
func (this *Recursor) Recurse(at *logic.Vector, deep int) {
	for _, d := range directions {
		<-this.Advance
		pos := at.Clone()
		pos.Translate(d)
		if pos.X >= 0 && pos.X < this.Arr.Size.X &&
			pos.Y >= 0 && pos.Y < this.Arr.Size.Y &&
			pos.Z >= 0 && pos.Z < this.Arr.Size.Z &&
			this.Arr.Content[pos.X][pos.Y][pos.Z] == -1 {
			this.Pieces++
			this.Arr.Content[pos.X][pos.Y][pos.Z] = this.ID
			this.Done <- true
			this.Recurse(pos, deep+1)
		} else {
			this.Done <- true
		}
	}

	if deep == 0 {
		this.Done <- false
	}
}
