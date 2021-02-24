package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type roomFunc func(w http.ResponseWriter, r *http.Request, room *room)

func (s *state) createRoom(w http.ResponseWriter, r *http.Request) {
	rm := &room{
		pz: puzzle{50, 20},
	}
	id := s.create(rm)

	http.Redirect(w, r, fmt.Sprintf("/room/%s", id), http.StatusFound)
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
		rf(w, r, room)
	})
}

func handleRoom(w http.ResponseWriter, r *http.Request, rm *room) {
	fmt.Fprintf(w, "Puzzle with dimensions %d x %d", rm.pz.width, rm.pz.height)
}

func handleRoomWS(w http.ResponseWriter, r *http.Request, rm *room) {
	wsc, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "could not upgrade connection", http.StatusInternalServerError)
		return
	}
	for {
		t, p, err := wsc.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		if err := wsc.WriteMessage(t, p); err != nil {
			log.Println(err)
			return
		}
	}
}
