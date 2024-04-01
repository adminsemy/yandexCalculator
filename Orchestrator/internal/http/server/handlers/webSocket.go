package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/web_socket/client"
	"github.com/gorilla/websocket"
)

type Manager struct {
	clients   map[*client.WebSocketClient]bool
	delete    chan *client.WebSocketClient
	messageCh chan *client.Message
	sync.RWMutex
}

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		clients:   make(map[*client.WebSocketClient]bool),
		delete:    make(chan *client.WebSocketClient),
		messageCh: make(chan *client.Message),
	}
	go m.ReadMessage()
	return m
}

func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {
	slog.Info("Установлено соединение с клиентом")
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Не удалось обновить соединение", err)
	}
	client := client.NewWebSocketClient(conn)
	m.addClient(client)
	slog.Info("Клиент добавлен", "клиент", *client)
	go client.ReadMessages(m.delete, m.messageCh)
	go client.WriteMessage(m.delete)
	go m.RemoveClient()
	client.WriteChan <- []byte("Hi from WS server")
}

func (m *Manager) addClient(client *client.WebSocketClient) {
	m.Lock()
	m.clients[client] = true
	m.Unlock()
}

func (m *Manager) RemoveClient() {
	var client *client.WebSocketClient
	for {
		client = <-m.delete
		m.Lock()
		delete(m.clients, client)
		m.Unlock()
		slog.Info("Client removed", "client", *client)
	}
}

func (m *Manager) ReadMessage() {
	for {
	}
}
