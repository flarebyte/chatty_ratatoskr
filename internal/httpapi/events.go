package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"sync"

	"github.com/coder/websocket"
)

type EventsAPI struct {
	allowedRoots map[string]struct{}
}

type clientMessage struct {
	ID       string   `json:"id,omitempty"`
	Kind     string   `json:"kind"`
	RootKeys []string `json:"rootKeys,omitempty"`
}

type serverMessage struct {
	ID       string   `json:"id,omitempty"`
	Kind     string   `json:"kind"`
	RootKeys []string `json:"rootKeys,omitempty"`
	Status   string   `json:"status,omitempty"`
	Message  string   `json:"message,omitempty"`
}

type connectionState struct {
	mu    sync.Mutex
	roots map[string]struct{}
}

func NewEventsAPI(allowedRoots []string) *EventsAPI {
	set := make(map[string]struct{}, len(allowedRoots))
	for _, root := range allowedRoots {
		set[root] = struct{}{}
	}
	return &EventsAPI{allowedRoots: set}
}

func (api *EventsAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("/events", api.handleEvents)
}

func (api *EventsAPI) handleEvents(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer func() {
		_ = conn.CloseNow()
	}()

	state := &connectionState{roots: map[string]struct{}{}}
	ctx := r.Context()

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return
		}

		var msg clientMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			_ = writeWS(ctx, conn, serverMessage{
				Kind:    "status",
				Status:  "invalid",
				Message: "invalid JSON payload",
			})
			continue
		}

		reply := api.handleMessage(state, msg)
		if err := writeWS(ctx, conn, reply); err != nil {
			return
		}
	}
}

func (api *EventsAPI) handleMessage(state *connectionState, msg clientMessage) serverMessage {
	switch msg.Kind {
	case "subscribe":
		if invalid := api.firstInvalidRoot(msg.RootKeys); invalid != "" {
			return serverMessage{
				ID:      msg.ID,
				Kind:    "status",
				Status:  "invalid",
				Message: "invalid rootKey: not in predefined allowed set",
			}
		}
		state.mu.Lock()
		for _, root := range msg.RootKeys {
			state.roots[root] = struct{}{}
		}
		roots := sortedRoots(state.roots)
		state.mu.Unlock()
		return serverMessage{ID: msg.ID, Kind: "subscribed", RootKeys: roots}
	case "unsubscribe":
		if invalid := api.firstInvalidRoot(msg.RootKeys); invalid != "" {
			return serverMessage{
				ID:      msg.ID,
				Kind:    "status",
				Status:  "invalid",
				Message: "invalid rootKey: not in predefined allowed set",
			}
		}
		state.mu.Lock()
		for _, root := range msg.RootKeys {
			delete(state.roots, root)
		}
		roots := sortedRoots(state.roots)
		state.mu.Unlock()
		return serverMessage{ID: msg.ID, Kind: "unsubscribed", RootKeys: roots}
	case "ping":
		return serverMessage{ID: msg.ID, Kind: "pong"}
	default:
		return serverMessage{
			ID:      msg.ID,
			Kind:    "status",
			Status:  "invalid",
			Message: "invalid command kind",
		}
	}
}

func (api *EventsAPI) firstInvalidRoot(roots []string) string {
	for _, root := range roots {
		if _, ok := api.allowedRoots[root]; !ok {
			return root
		}
	}
	return ""
}

func sortedRoots(roots map[string]struct{}) []string {
	out := make([]string, 0, len(roots))
	for root := range roots {
		out = append(out, root)
	}
	sort.Strings(out)
	return out
}

func writeWS(ctx context.Context, conn *websocket.Conn, msg serverMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, data)
}
