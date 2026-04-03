package controller

import (
	"erb-backend/src/broadcaster"
	"net/http"
)

type Broadcaster interface {
	Subscribe() chan string
	Unsubscribe(chan string)
	Publish(broadcaster.Event)
}

type EventsStreamController struct {
	broadcaster *broadcaster.Broadcaster
}

func NewEventsStreamController(broadcaster *broadcaster.Broadcaster) *EventsStreamController {
	return &EventsStreamController{broadcaster: broadcaster}
}

// ServeHTTP godoc
// @Summary     Events Stream
// @Description Server-sent events stream for real-time wagon movement updates
// @Tags        Events
// @Produce     text/event-stream
// @Success     200  {string}  string  "SSE stream"
// @Router      /events/stream [get]
func (h *EventsStreamController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	ch := h.broadcaster.Subscribe()
	defer h.broadcaster.Unsubscribe(ch)

	for {
		select {
		case msg := <-ch:
			_, err := w.Write([]byte(msg + "\n"))
			if err != nil {
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
