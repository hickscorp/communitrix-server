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
	ats := make(logic.Vectors, 0, count)
	for i := 0; i < count; i++ {
		at := logic.NewVectorFromInts(rand.Intn(piece.Size.X), rand.Intn(piece.Size.Y), rand.Intn(piece.Size.Z))
		if at.SameAsAny(ats) {
			i--
			continue
		}
		ats = append(ats, at)
	}

	// Prepare our pieces array to return.
	ret := make(logic.Pieces, count)
	for i := 0; i < count; i++ {
		ret[i] = logic.NewPiece(logic.NullVector, 0)
	}
	// Start as many goroutines as we have starting points.
	recursors := make(Recursors, count)
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
		if recursor != nil {
			recursor.Advance <- true
			if cell := <-recursor.Done; cell != nil {
				piece := ret[idx]
				piece.Content = append(piece.Content, cell)
			} else {
				log.Warning("Got Recursor %d end!", recursor.ID)
				recursors[idx] = nil
				done++
			}
		}
	}
	for _, piece := range ret {
		piece.CleanUp()
	}
	return ret, true
}

type Recursor struct {
	ID      int
	Arr     *array.ContentArray
	Advance chan bool
	Done    chan *logic.Cell
	Pieces  int
}
type Recursors []*Recursor

func NewRecursor(id int, arr *array.ContentArray) *Recursor {
	return &Recursor{
		ID:      id,
		Arr:     arr,
		Advance: make(chan bool),
		Done:    make(chan *logic.Cell),
		Pieces:  1,
	}
}
func (this *Recursor) Recurse(at *logic.Vector, depth int) {
	dir := shuffle(directions.Clone())
	<-this.Advance
	for _, d := range dir {
		pos := at.Clone()
		pos.Translate(d)
		if pos.X >= 0 && pos.X < this.Arr.Size.X && pos.Y >= 0 && pos.Y < this.Arr.Size.Y && pos.Z >= 0 && pos.Z < this.Arr.Size.Z && this.Arr.Content[pos.X][pos.Y][pos.Z] == -1 {
			this.Pieces++
			this.Arr.Content[pos.X][pos.Y][pos.Z] = this.ID
			this.Done <- logic.NewCellFromInts(pos.X, pos.Y, pos.Z, this.ID)
			this.Recurse(pos, depth+1)
		}
	}

	if depth == 0 {
		log.Warning("Recursor %d: Signaling stop.", this.ID)
		this.Done <- nil
	}
}

func shuffle(arr logic.Vectors) logic.Vectors {
	for i := len(arr) - 1; i > 0; i-- {
		j := rand.Intn(i)
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}