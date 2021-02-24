package main

import (
	"fmt"
)

type shape int

const (
	square shape = iota
	circle
	triangle
)

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// a chain is the set of pieces that have fit together.
// chains can add themselves to other chains, and return their pieces.
type chain interface {
	// add returns a slice of pieces with chain added to c.
	add(c chain) pieces
	pieces() pieces
}

// piece is a puzzle piece, which is a chain of length 1
type piece struct {
	x, y int
	// sides [4]shape
}

func (p piece) fits(t piece) bool {
	x := abs(p.x-t.x) == 1
	y := abs(p.y-t.y) == 1
	return (x || y) && !(x && y) // XOR
}

func (p piece) add(c chain) pieces {
	ps := c.pieces()
	for _, t := range ps {
		if p.fits(t) {
			return append(ps, p)
		}
	}
	return ps
}

func (p piece) pieces() pieces {
	return []piece{p}
}

func (p piece) String() string {
	return fmt.Sprintf("Piece at (%d, %d)", p.x, p.y)
}

type pieces []piece

func (ps pieces) add(c chain) pieces {
	tps := c.pieces()
	for _, p := range ps {
		for _, t := range tps {
			if p.fits(t) {
				return append(tps, ps...)
			}
		}
	}
	return tps
}

func (ps pieces) pieces() pieces { return ps }

// a puzzle represents a configured puzzle, with a
// width and height.
type puzzle struct {
	width, height int
}

func (pz *puzzle) solvedBy(c chain) bool {
	return len(c.pieces()) == pz.height*pz.width
}
