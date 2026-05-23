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
