package database

import (
	"context"
	"database/sql"
	"time"

	"flow/internal/models"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Conversation methods

func (r *Repository) CreateConversation(ctx context.Context, title string) (*models.Conversation, error) {
	id := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO conversations (id, title, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, created_at, updated_at
	`

	conv := &models.Conversation{}
	err := r.db.QueryRowContext(ctx, query, id, title, now, now).Scan(
		&conv.ID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	conv.Messages = make([]*models.Message, 0)
	return conv, nil
}

func (r *Repository) GetConversation(ctx context.Context, id string) (*models.Conversation, error) {
	query := `
		SELECT id, title, created_at, updated_at
		FROM conversations
		WHERE id = $1
	`

	conv := &models.Conversation{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&conv.ID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	messages, err := r.GetMessages(ctx, id)
	if err != nil {
		return nil, err
	}
	conv.Messages = messages

	return conv, nil
}

func (r *Repository) ListConversations(ctx context.Context, limit, offset int) ([]*models.Conversation, error) {
	query := `
		SELECT id, title, created_at, updated_at
		FROM conversations
		ORDER BY updated_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []*models.Conversation
	for rows.Next() {
		conv := &models.Conversation{}
		err := rows.Scan(&conv.ID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt)
		if err != nil {
			continue
		}
		convs = append(convs, conv)
	}

	return convs, nil
}

func (r *Repository) DeleteConversation(ctx context.Context, id string) error {
	query := `DELETE FROM conversations WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Message methods

func (r *Repository) CreateMessage(ctx context.Context, conversationID string, msg *models.Message) error {
	query := `
		INSERT INTO messages (id, conversation_id, role, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query,
		msg.ID, conversationID, msg.Role, msg.Content, msg.Timestamp,
	)
	return err
}

func (r *Repository) GetMessages(ctx context.Context, conversationID string) ([]*models.Message, error) {
	query := `
		SELECT id, role, content, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		msg := &models.Message{}
		var roleStr string
		err := rows.Scan(&msg.ID, &roleStr, &msg.Content, &msg.Timestamp)
		if err != nil {
			continue
		}
		msg.Role = models.Role(roleStr)
		messages = append(messages, msg)
	}

	return messages, nil
}

// Email methods

type Email struct {
	ID           string
	MessageID    string
	Subject      string
	Sender       string
	Recipients   []string
	Body         string
	SentDate     time.Time
	GraphNodeID  int64
}

func (r *Repository) CreateEmail(ctx context.Context, email *Email) error {
	query := `
		INSERT INTO emails (id, message_id, subject, sender, recipients, body, sent_date, graph_node_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	id := uuid.New().String()
	_, err := r.db.ExecContext(ctx, query,
		id, email.MessageID, email.Subject, email.Sender,
		email.Recipients, email.Body, email.SentDate, email.GraphNodeID,
	)
	return err
}

func (r *Repository) SearchEmails(ctx context.Context, searchTerm string, limit int) ([]*Email, error) {
	query := `
		SELECT id, message_id, subject, sender, body, sent_date, graph_node_id
		FROM emails
		WHERE subject ILIKE $1 OR body ILIKE $1 OR sender ILIKE $1
		ORDER BY sent_date DESC
		LIMIT $2
	`

	searchPattern := "%" + searchTerm + "%"
	rows, err := r.db.QueryContext(ctx, query, searchPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []*Email
	for rows.Next() {
		email := &Email{}
		err := rows.Scan(
			&email.ID, &email.MessageID, &email.Subject,
			&email.Sender, &email.Body, &email.SentDate, &email.GraphNodeID,
		)
		if err != nil {
			continue
		}
		emails = append(emails, email)
	}

	return emails, nil
}

func (r *Repository) GetEmailsBySender(ctx context.Context, sender string, limit int) ([]*Email, error) {
	query := `
		SELECT id, message_id, subject, sender, body, sent_date, graph_node_id
		FROM emails
		WHERE sender = $1
		ORDER BY sent_date DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, sender, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []*Email
	for rows.Next() {
		email := &Email{}
		err := rows.Scan(
			&email.ID, &email.MessageID, &email.Subject,
			&email.Sender, &email.Body, &email.SentDate, &email.GraphNodeID,
		)
		if err != nil {
			continue
		}
		emails = append(emails, email)
	}

	return emails, nil
}

// Stats

type DBStats struct {
	ConversationCount int64
	MessageCount      int64
	EmailCount        int64
	EntityCount       int64
}

func (r *Repository) GetStats(ctx context.Context) (*DBStats, error) {
	stats := &DBStats{}

	queries := []struct {
		query string
		dest  *int64
	}{
		{"SELECT COUNT(*) FROM conversations", &stats.ConversationCount},
		{"SELECT COUNT(*) FROM messages", &stats.MessageCount},
		{"SELECT COUNT(*) FROM emails", &stats.EmailCount},
		{"SELECT COUNT(*) FROM entities", &stats.EntityCount},
	}

	for _, q := range queries {
		r.db.QueryRowContext(ctx, q.query).Scan(q.dest)
	}

	return stats, nil
}
