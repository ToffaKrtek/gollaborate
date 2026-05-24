package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestIndexPage_ShowsDocumentName(t *testing.T) {
	req := httptest.NewRequest("GET", "/?doc=main.go", nil)
	rec := httptest.NewRecorder()
	http.HandlerFunc(indexHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()

	assert.Contains(t, body, "<title>Редактор: main.go</title>")
	assert.Contains(t, body, `data-doc="main.go"`)
	assert.Contains(t, body, `window.currentDoc = "main.go";`)
}

func TestIndexPage_OfflineStructure(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	http.HandlerFunc(indexHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()

	assert.Contains(t, body, `id="code-editor"`)
	assert.Contains(t, body, `spellcheck="false"`)
	assert.Contains(t, body, `<textarea`)
	assert.Contains(t, body, `id="status"`)

	assert.Contains(t, body, `href="/static/editor.css"`)
	assert.Contains(t, body, `src="/static/editor.js"`)

	assert.NotContains(t, body, "https://")
	assert.NotContains(t, body, "//cdnjs")
	assert.NotContains(t, body, "unpkg")
	assert.NotContains(t, body, "cdn.jsdelivr")
}

func TestWebSocketRooms(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wsHandler(w, r)
	}))
	defer server.Close()
	wsUrl := "ws" + strings.TrimPrefix(server.URL, "http")
	conn1, _, err := websocket.DefaultDialer.Dial(wsUrl+"?doc=doc1", nil)
	assert.NoError(t, err)
	defer conn1.Close()
	conn2, _, err := websocket.DefaultDialer.Dial(wsUrl+"?doc=doc1", nil)
	assert.NoError(t, err)
	defer conn2.Close()
	conn3, _, err := websocket.DefaultDialer.Dial(wsUrl+"?doc=doc2", nil)
	assert.NoError(t, err)
	defer conn3.Close()

	msg1 := "hello from client1"
	err = conn1.WriteMessage(websocket.TextMessage, []byte(msg1))
	assert.NoError(t, err)

	_, msg2, err := conn2.ReadMessage()
	assert.NoError(t, err)
	assert.Equal(t, msg1, string(msg2))

	conn3.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, _, err = conn3.ReadMessage()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestWebSocketEcho(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(wsEchoHandler))
	defer server.Close()

	wsUrl := "ws" + strings.TrimPrefix(server.URL, "http")

	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	assert.NoError(t, err)
	defer conn.Close()

	testMsg := "hello, echo!"
	err = conn.WriteMessage(websocket.TextMessage, []byte(testMsg))
	assert.NoError(t, err)

	_, msg, err := conn.ReadMessage()
	assert.NoError(t, err)
	assert.Equal(t, testMsg, string(msg))
}

func TestIndexPage(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(indexHandler)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, rec.Code, http.StatusOK)
	body := rec.Body.String()
	assert.Contains(t, body, "<div")
}
