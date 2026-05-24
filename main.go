package main

import (
	"embed"
	"html/template"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gorilla/websocket"
)

var roomManager = NewRoomManager()

type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string]map[*websocket.Conn]struct{}
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]map[*websocket.Conn]struct{}),
	}
}

func (rm *RoomManager) Join(doc string, conn *websocket.Conn) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if rm.rooms[doc] == nil {
		rm.rooms[doc] = make(map[*websocket.Conn]struct{})
	}
	rm.rooms[doc][conn] = struct{}{}
}

func (rm *RoomManager) Leave(doc string, conn *websocket.Conn) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, ok := rm.rooms[doc]; ok {
		delete(rm.rooms[doc], conn)
		if len(rm.rooms[doc]) == 0 {
			delete(rm.rooms, doc)
		}
	}
}

func (rm *RoomManager) Broadcast(doc string, msgType int, msg []byte, sender *websocket.Conn) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	for conn := range rm.rooms[doc] {
		if conn == sender {
			continue
		}
		conn.WriteMessage(msgType, msg)
	}
}

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
	roomManager.Join(doc, conn)

	defer roomManager.Leave(doc, conn)
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		roomManager.Broadcast(doc, msgType, msg, conn)
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

//go:embed static/*
var staticFiles embed.FS
var indexTemplate = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Редактор: {{.Doc}}</title>
	<link rel="stylesheet" href="/static/editor.css">
</head>
<body>
	<div id="status" class="status" data-doc="{{.Doc}}">Disconnected</div>
	<textarea id="code-editor" spellcheck="false" placeholder="Начните редактирование..."></textarea>
	<script>
		// .Doc уже экранирован в Go, можно вставлять напрямую в строку JS
		window.currentDoc = "{{.Doc}}";
	</script>
	<script src="/static/editor.js" defer></script>
</body>
</html>`))

func setContentType(w http.ResponseWriter, path string) {
	ext := filepath.Ext(path)
	switch ext {
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".html", ".htm":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/static/"):]
	data, err := staticFiles.ReadFile("static/" + path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	setContentType(w, path)
	w.Write(data)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	doc := r.URL.Query().Get("doc")
	if doc == "" {
		doc = "untitled"
	}

	escapeDoc := template.JSEscapeString(doc)

	data := struct {
		Doc string
	}{
		Doc: escapeDoc,
	}

	if err := indexTemplate.Execute(w, data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func main() {
	// http.HandleFunc("/ws", wsEchoHandler)
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/static/", staticHandler)
	http.HandleFunc("/", indexHandler)

	http.ListenAndServe("0.0.0.0:8080", nil)
}
