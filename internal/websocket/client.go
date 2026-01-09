package websocket

import (
	"encoding/json"
	"time"

	"flow/internal/models"
	"flow/pkg/logger"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8192
)

type Client struct {
	ID             string
	hub            *Hub
	conn           *websocket.Conn
	send           chan []byte
	log            *logger.Logger
	messageHandler func(*Client, []byte)
}

func NewClient(hub *Hub, conn *websocket.Conn, log *logger.Logger, handler func(*Client, []byte)) *Client {
	return &Client{
		ID:             uuid.New().String(),
		hub:            hub,
		conn:           conn,
		send:           make(chan []byte, 256),
		log:            log.WithComponent("websocket-client"),
		messageHandler: handler,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.log.Error().Err(err).Msg("Unexpected close")
			}
			break
		}

		if c.messageHandler != nil {
			c.messageHandler(c, message)
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) SendEvent(event *models.StreamEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	c.send <- data
	return nil
}

func (c *Client) SendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.send <- data
	return nil
}
