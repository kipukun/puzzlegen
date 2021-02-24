package main

import (
	"context"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// a room is a hub for clients.
// the room manages communication between clients.
type room struct {
	id   string
	pz   puzzle
	in   chan string
	outs map[string]chan string
	mu   sync.RWMutex
}

func (r *room) request() (string, <-chan string) {
	out := make(chan string, 1)
	id := id()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.outs[id] = out
	return id, out
}

func (r *room) done(i string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.outs, i)
}

func (r *room) close() {
	r.mu.RLock()
	for _, out := range r.outs {
		close(out)
	}
	r.mu.RUnlock()
}

func (r *room) broadcast(ctx context.Context) {
	for {
		select {
		case msg := <-r.in:
			r.mu.RLock()
			for _, out := range r.outs {
				out <- msg
			}
			r.mu.RUnlock()
		case <-ctx.Done():
			r.close()
			return
		}
	}
}

type state struct {
	srv   *http.Server
	rooms map[string]*room
	mu    sync.RWMutex
}

func (s *state) create(ctx context.Context) *room {
	r := new(room)
	r.in = make(chan string)
	r.outs = make(map[string]chan string)
	r.pz = puzzle{50, 20}
	i := id()
	r.id = i
	s.mu.Lock()
	s.rooms[i] = r
	s.mu.Unlock()

	go r.broadcast(ctx)

	return r
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
		tmpl, err := template.ParseFiles("static/main.html")
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	r.Handle("/create", s.NeedsAContext(ctx, s.createRoom)).Methods("POST")

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
	go func() {
		<-ctx.Done()
		fmt.Println("got interrupt, stopping server and closing rooms...")
		stop()
		s.srv.Close()
	}()
	s.srv.ListenAndServe()
}
