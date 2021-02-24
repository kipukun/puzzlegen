package main

import (
	"fmt"
	"testing"
)

func TestCreateRoom(t *testing.T) {
	s := new(state)
	s.rooms = make(map[string]*room)
	r := new(room)
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("create #%d", i), func(t *testing.T) {
			s.create(r)
		})
	}
}

func BenchmarkCreateRoom(b *testing.B) {
	s := new(state)
	s.rooms = make(map[string]*room)
	r := new(room)
	for i := 0; i < b.N; i++ {
		s.create(r)
	}
}
