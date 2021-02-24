package main

import (
	"html/template"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// a room is a hub for clients.
// the room manages communication between clients.
type room struct {
	pz puzzle
}

type state struct {
	srv   *http.Server
	rooms map[string]*room
	mu    sync.RWMutex
	done  <-chan bool
}

func (s *state) create(r *room) string {
	i := id()
	s.mu.Lock()
	s.rooms[i] = r
	s.mu.Unlock()

	return i
}

func (s *state) get(id string) *room {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rooms[id]
}

func main() {
	s := new(state)
	rand.Seed(time.Now().UnixNano())

	s.rooms = make(map[string]*room)

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ps := pieces{piece{0, 0}, piece{0, 1}, piece{1, 0}}
		tmpl, err := template.ParseFiles("main.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusTeapot)
			return
		}
		err = tmpl.Execute(w, ps)
		if err != nil {
			http.Error(w, err.Error(), http.StatusTeapot)
			return
		}

	})
	r.HandleFunc("/create", s.createRoom).Methods("POST")

	rooms := r.PathPrefix("/room").Subrouter()
	rooms.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusPermanentRedirect) // true /
	})

	rooms.Handle("/{id}", s.NeedsARoom(handleRoom))
	rooms.Handle("/{id}/relay", s.NeedsARoom(handleRoomWS))

	s.srv = &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: r,
	}
	s.srv.ListenAndServe()
}