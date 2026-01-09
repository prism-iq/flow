package models

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Messages  []*Message `json:"messages"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	mu        sync.RWMutex
}

func NewConversation() *Conversation {
	return &Conversation{
		ID:        uuid.New().String(),
		Title:     "New Conversation",
		Messages:  make([]*Message, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (c *Conversation) AddMessage(msg *Message) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Messages = append(c.Messages, msg)
	c.UpdatedAt = time.Now()
}

func (c *Conversation) GetMessages() []*Message {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Messages
}
