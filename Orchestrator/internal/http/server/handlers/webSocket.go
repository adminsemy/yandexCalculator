package handlers

import (
	"log/slog"
	"net/http"

	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/config"
	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/web_socket/client"
	"github.com/gorilla/websocket"
)

func serveWS(c *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Установлено соединение с клиентом ws")
		upgrade := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		conn, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("Не удалось обновить соединение ws", err)
		}
		client := client.NewWebSocketClient(conn, client.ClientExpression)
		slog.Info("Клиент ws добавлен", "клиент", *client)
		c.WSmanager.AddClient(client)
	}
}
