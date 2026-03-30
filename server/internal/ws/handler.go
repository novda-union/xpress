package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// UserWebSocket handles WebSocket connections for customer app users
func UserWebSocket(hub *Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, ok := c.Get("user_id").(string)
		if !ok || userID == "" {
			return echo.ErrUnauthorized
		}

		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Printf("ws upgrade error: %v", err)
			return err
		}

		client := NewClient(hub, conn)
		client.AutoSubscribe("user:" + userID)

		go client.WritePump()
		go client.ReadPump()

		return nil
	}
}

// AdminWebSocket handles WebSocket connections for store staff
func AdminWebSocket(hub *Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		storeID, ok := c.Get("store_id").(string)
		if !ok || storeID == "" {
			return echo.ErrUnauthorized
		}

		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Printf("ws upgrade error: %v", err)
			return err
		}

		client := NewClient(hub, conn)
		if branchID, ok := c.Get("branch_id").(*string); ok && branchID != nil && *branchID != "" {
			client.AutoSubscribe("branch:" + *branchID)
		} else {
			client.AutoSubscribe("store:" + storeID)
		}

		go client.WritePump()
		go client.ReadPump()

		return nil
	}
}
