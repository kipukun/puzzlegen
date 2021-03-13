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

// a chain is the set of pieces that have fit together.
type chain interface {
	fits(c chain) bool
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

func (p piece) pieces() pieces {
	return []piece{p}
}

func (p piece) String() string {
	return fmt.Sprintf("Piece at (%d, %d)", p.x, p.y)
}

type pieces []piece

func (ps pieces) fits(c chain) bool {
	tps := c.pieces()
	for _, p := range ps {
		for _, t := range tps {
			if p.fits(t) {
				return true
			}
		}
	}
	return false
}

func (ps pieces) pieces() pieces { return ps }
