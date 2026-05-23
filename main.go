package main

import (
	"html/template"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var rooms = struct {
	sync.RWMutex
	m map[string][]*websocket.Conn
}{m: make(map[string][]*websocket.Conn)}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	doc := r.URL.Query().Get("doc")
	if doc == "" {
		http.Error(w, "missing doc parameter", http.StatusBadRequest)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "could no upgrade", http.StatusBadRequest)
		return
	}
	rooms.Lock()
	rooms.m[doc] = append(rooms.m[doc], conn)
	rooms.Unlock()

	defer func() {
		rooms.Lock()
		defer rooms.Unlock()
		conns := rooms.m[doc]
		for i, c := range conns {
			if c == conn {
				rooms.m[doc] = append(rooms.m[doc][:i], rooms.m[doc][i+1:]...)
				break
			}
		}
		if len(rooms.m[doc]) == 0 {
			delete(rooms.m, doc)
		}
		conn.Close()
	}()
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		rooms.RLock()
		for _, c := range rooms.m[doc] {
			if c == conn {
				continue
			}
			if err := c.WriteMessage(msgType, msg); err != nil {
				c.Close()
			}
		}
		rooms.RUnlock()
	}
}

func wsEchoHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "could no upgrade", http.StatusBadRequest)
	}
	defer conn.Close()

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if err := conn.WriteMessage(msgType, msg); err != nil {
			break
		}
	}
}

var indexTemplate = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html>
<head>
    <title>Collaborative Code Editor</title>
</head>
<body>
    <div id="editor"></div>
</body>
</html>`))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	indexTemplate.Execute(w, nil)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/ws", wsEchoHandler)
	http.HandleFunc("/ws", wsHandler)
	http.ListenAndServe("0.0.0.0:8080", nil)
}
