package main

import (
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
)

func TestHandleRoomWs(t *testing.T) {
	s := new(state)
	ts := httptest.NewServer(s.NeedsARoom(handleRoomWS))

	c, _, err := websocket.DefaultDialer.Dial(ts.URL, nil)
	if err != nil {
		t.Fatal(err)
		return
	}
}
