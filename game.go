package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"image"
	"image/draw"
	"image/jpeg"
	"io"
	"sort"
	"time"
)

type msg struct {
	src, targ pieces
}

type res struct {
	m       *msg
	success bool
}

// game represents a game, in a room.
// a game contains when the game started, the underlying image,
// and the puzzle to be solved.
type game struct {
	img     *image.RGBA
	created time.Time
	// number of pieces in the puzzle
	nX, nY int
	// size of each piece, in pixels
	sW, sH int
}

func (g *game) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	jenc := json.NewEncoder(&b)
	err := jpeg.Encode(&b, g.img, nil)
	if err != nil {
		return []byte{}, err
	}
	// encoded := base64.StdEncoding.EncodeToString(b.Bytes())
	b.Reset()
	d := struct {
		Created        time.Time
		NX, NY, SW, SH int
	}{
		g.created,
		g.nX, g.nY, g.sW, g.sH,
	}
	err = jenc.Encode(d)
	if err != nil {
		return []byte{}, err
	}
	return b.Bytes(), nil
}

// returns an initalized game.
// newGame expects r to contain valid image data.
func newGame(r io.Reader, n int) (*game, error) {
	var buf bytes.Buffer
	g := new(game)
	io.Copy(&buf, r)
	m, _, err := image.Decode(&buf)
	if err != nil {
		return nil, err
	}

	diff := n
	factors := make([]int, 2)

	for i := 1; i < n/2; i++ {
		if n/2%i == 0 {
			if abs(i-n/i) < diff {
				diff = abs(i - n/i)
				factors[0] = i
				factors[1] = n / i
			}
		}
	}
	sort.Ints(factors)

	width := m.Bounds().Dx()
	height := m.Bounds().Dy()

	if width > height {
		g.sW = width / (factors[1] - 1)
		g.sH = height / (factors[0] - 1)
	} else {
		g.sH = height / (factors[1] - 1)
		g.sW = width / (factors[0] - 1)
	}

	g.nX = width / g.sW
	g.nY = height / g.sH

	// https://blog.golang.org/image-draw#TOC_6.
	b := m.Bounds()
	img := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(img, m.Bounds(), m, b.Min, draw.Src)

	g.img = img
	g.created = time.Now()

	return g, nil
}

func (g *game) imageAt(x, y int) (image.Image, error) {
	if x < 1 || y < 1 || y > g.nY || x > g.nX {
		return nil, errors.New("range out of bounds")
	}
	maxx := x * g.sW
	maxy := y * g.sH
	return g.img.SubImage(image.Rect(maxx-g.sW, maxy-g.sH, maxx, maxy)), nil
}
