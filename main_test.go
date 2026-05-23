package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

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
