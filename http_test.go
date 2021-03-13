package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func tows(s string) string { return strings.Replace(s, "http", "ws", 1) }

type testRoomWS struct {
	r *room
}

func (trws *testRoomWS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handleRoomWS(trws.r, w, r)
}

func TestHandleRoomWs(t *testing.T) {
	num := 50
	s := new(state)
	s.rooms = make(map[string]*room)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := s.create(ctx, nil)
	ts := httptest.NewServer(&testRoomWS{
		r: r,
	})

	ch := make(chan []byte, num)

	parent, _, err := websocket.DefaultDialer.Dial(tows(ts.URL), nil)
	if err != nil {
		t.Fatalf("%v", err)
		return
	}
	for i := 0; i < num; i++ {
		t.Run(fmt.Sprintf("client #%d", i), func(t *testing.T) {
			conn, _, err := websocket.DefaultDialer.Dial(tows(ts.URL), nil)
			if err != nil {
				t.Fatalf("%v", err)
				return
			}
			msg := []byte(fmt.Sprintf("hello from client %d", i))
			conn.WriteMessage(websocket.TextMessage, msg)
			_, p, err := parent.ReadMessage()
			if err != nil {
				t.Fatalf("%v", err)
				return
			}
			ch <- p
		})
	}
	if len(ch) != cap(ch) {
		t.Fatalf("did not fill channel %d/%d", len(ch), cap(ch))
	}
}
