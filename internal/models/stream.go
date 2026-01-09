package models

type StreamEvent struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
	Done    bool   `json:"done,omitempty"`
}

const (
	EventTypeToken = "token"
	EventTypeDone  = "done"
	EventTypeError = "error"
	EventTypeStart = "start"
)

func NewTokenEvent(content string) *StreamEvent {
	return &StreamEvent{
		Type:    EventTypeToken,
		Content: content,
	}
}

func NewDoneEvent() *StreamEvent {
	return &StreamEvent{
		Type: EventTypeDone,
		Done: true,
	}
}

func NewErrorEvent(err string) *StreamEvent {
	return &StreamEvent{
		Type:  EventTypeError,
		Error: err,
	}
}

func NewStartEvent() *StreamEvent {
	return &StreamEvent{
		Type: EventTypeStart,
	}
}
