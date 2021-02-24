package main

import (
	"context"
	"fmt"
	"testing"
)

func TestCreateRoom(t *testing.T) {
	s := new(state)
	s.rooms = make(map[string]*room)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("create #%d", i), func(t *testing.T) {
			s.create(ctx)
		})
	}
}

func BenchmarkCreateRoom(b *testing.B) {
	s := new(state)
	s.rooms = make(map[string]*room)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < b.N; i++ {
		s.create(ctx)
	}
}
