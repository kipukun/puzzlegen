package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type roomFunc func(rm *room, w http.ResponseWriter, r *http.Request)
type createFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (s *state) createRoom(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	rm := s.create(ctx)

	http.Redirect(w, r, fmt.Sprintf("/room/%s", rm.id), http.StatusFound)
}

func (s *state) NeedsAContext(ctx context.Context, cf createFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cf(ctx, w, r)
	})
}

func (s *state) NeedsARoom(rf roomFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v, ok := mux.Vars(r)["id"]
		if !ok {
			http.Error(w, "no id given", http.StatusNotFound)
			return
		}
		room := s.get(v)
		if room == nil {
			http.Error(w, "room not found", http.StatusNotFound)
			return
		}
		rf(room, w, r)
	})
}

func handleRoom(rm *room, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Puzzle with dimensions %d x %d", rm.pz.width, rm.pz.height)
}

func handleRoomWS(rm *room, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "could not upgrade connection", http.StatusInternalServerError)
		return
	}
	recv := rm.request()
	go func() {
		for msg := range recv {
			conn.WriteMessage(websocket.TextMessage, []byte(msg))
		}
	}()
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
		}
		rm.in <- string(p)
	}
}
