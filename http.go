package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type roomFunc func(rm *room, w http.ResponseWriter, r *http.Request)
type createFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
}

func (s *state) createRoom(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // 10MB
	f, _, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
		http.Error(w, "could not parse your file", http.StatusInternalServerError)
		return
	}

	g, err := newGame(f, 100)
	if err != nil {
		log.Println(err)
		return
	}
	f.Close()

	rm := s.create(ctx, g)

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
		X, Y int
	}{
		rm.id,
		rm.g.nX, rm.g.nY,
	}
	err = tmpl.Execute(w, d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusTeapot)
		return
	}
}

func handleGameInfo(rm *room, w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(rm.g)
	if err != nil {
		log.Println(err)
		return
	}
	w.Write(b)
}

func handleGetImage(rm *room, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age:290304000, public")
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Last-Modified", time.Now().Format(http.TimeFormat))
	var b bytes.Buffer
	vars := mux.Vars(r)
	x, _ := strconv.Atoi(vars["x"])
	y, _ := strconv.Atoi(vars["y"])

	img, err := rm.g.imageAt(x, y)
	if err != nil {
		log.Println(err)
		return
	}
	err = jpeg.Encode(&b, img, nil)
	if err != nil {
		log.Println(err)
		return
	}
	io.Copy(w, &b)
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
