package main

import (
	"context"
	"fmt"
	"html/template"
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
	tmpl, err := template.ParseFiles("static/room.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusTeapot)
		return
	}
	d := struct {
		Name string
	}{
		rm.id,
	}
	err = tmpl.Execute(w, d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusTeapot)
		return
	}
}

func handleRoomWS(rm *room, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "could not upgrade connection", http.StatusInternalServerError)
		return
	}
	id, recv := rm.request()
	done := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case msg := <-recv:
				conn.WriteMessage(websocket.TextMessage, []byte(msg))
			case <-done:
				log.Println("got done!")
				rm.done(id)
				return
			}
		}
	}()
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v, user-agent: %v", err, r.Header.Get("User-Agent"))
			}
			fmt.Println("done with err:", err)
			done <- struct{}{}
			return
		}
		rm.in <- string(p)
	}
}
