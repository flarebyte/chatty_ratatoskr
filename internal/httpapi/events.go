package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type EventsAPI struct {
	allowedRoots      map[string]struct{}
	generateID        func() string
	now               func() string
	readLimitBytes    int64
	heartbeatInterval time.Duration
	heartbeatTimeout  time.Duration
	mu                sync.Mutex
	subscribers       map[*subscriber]struct{}
}

type clientMessage struct {
	ID       string   `json:"id,omitempty"`
	Kind     string   `json:"kind"`
	RootKeys []string `json:"rootKeys,omitempty"`
}

type serverMessage struct {
	ID       string         `json:"id,omitempty"`
	Kind     string         `json:"kind"`
	RootKeys []string       `json:"rootKeys,omitempty"`
	Status   string         `json:"status,omitempty"`
	Message  string         `json:"message,omitempty"`
	Event    *eventEnvelope `json:"event,omitempty"`
}

type connectionState struct {
	mu    sync.Mutex
	roots map[string]struct{}
}

type subscriber struct {
	conn  *websocket.Conn
	state *connectionState
}

type eventEnvelope struct {
	EventID         string          `json:"eventId"`
	RootKey         keyParams       `json:"rootKey"`
	Operation       string          `json:"operation"`
	Created         string          `json:"created"`
	Key             *keyParams      `json:"key,omitempty"`
	KeyValue        *keyValueParams `json:"keyValue,omitempty"`
	SnapshotVersion string          `json:"snapshotVersion,omitempty"`
}

func NewEventsAPI(allowedRoots []string) *EventsAPI {
	set := make(map[string]struct{}, len(allowedRoots))
	for _, root := range allowedRoots {
		set[root] = struct{}{}
	}
	return &EventsAPI{
		allowedRoots:      set,
		generateID:        func() string { return "event-generated" },
		now:               func() string { return time.Now().UTC().Format(time.RFC3339) },
		readLimitBytes:    32768,
		heartbeatInterval: 30 * time.Second,
		heartbeatTimeout:  5 * time.Second,
		subscribers:       map[*subscriber]struct{}{},
	}
}

func NewEventsAPIWithOptions(allowedRoots []string, generateID func() string, now func() string) *EventsAPI {
	api := NewEventsAPI(allowedRoots)
	if generateID != nil {
		api.generateID = generateID
	}
	if now != nil {
		api.now = now
	}
	return api
}

func NewEventsAPIWithRuntimeOptions(allowedRoots []string, generateID func() string, now func() string, readLimitBytes int64, heartbeatInterval, heartbeatTimeout time.Duration) *EventsAPI {
	api := NewEventsAPIWithOptions(allowedRoots, generateID, now)
	if readLimitBytes > 0 {
		api.readLimitBytes = readLimitBytes
	}
	if heartbeatInterval > 0 {
		api.heartbeatInterval = heartbeatInterval
	}
	if heartbeatTimeout > 0 {
		api.heartbeatTimeout = heartbeatTimeout
	}
	return api
}

func (api *EventsAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("/events", api.handleEvents)
}

func (api *EventsAPI) handleEvents(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	conn.SetReadLimit(api.readLimitBytes)
	defer func() {
		_ = conn.CloseNow()
	}()

	state := &connectionState{roots: map[string]struct{}{}}
	sub := &subscriber{conn: conn, state: state}
	api.addSubscriber(sub)
	defer api.removeSubscriber(sub)
	ctx := r.Context()
	go api.runHeartbeat(ctx, conn)

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

func (api *EventsAPI) runHeartbeat(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(api.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(context.Background(), api.heartbeatTimeout)
			err := conn.Ping(pingCtx)
			cancel()
			if err != nil {
				_ = conn.Close(websocket.StatusPolicyViolation, "heartbeat failed")
				return
			}
		}
	}
}

func (api *EventsAPI) EmitSet(rootKey keyParams, keyValue keyValueParams) {
	keyCopy := keyValue.Key
	keyValueCopy := keyValue
	api.emit(eventEnvelope{
		EventID:   api.generateID(),
		RootKey:   rootKey,
		Operation: "set",
		Created:   api.now(),
		Key:       &keyCopy,
		KeyValue:  &keyValueCopy,
	})
}

func (api *EventsAPI) EmitSnapshotReplaced(rootKey keyParams, snapshotVersion string) {
	api.emit(eventEnvelope{
		EventID:         api.generateID(),
		RootKey:         rootKey,
		Operation:       "snapshot-replaced",
		Created:         api.now(),
		SnapshotVersion: snapshotVersion,
	})
}

func (api *EventsAPI) handleMessage(state *connectionState, msg clientMessage) serverMessage {
	switch msg.Kind {
	case "subscribe":
		if invalid := api.firstInvalidRoot(msg.RootKeys); invalid != "" {
			return invalidRootStatusMessage(msg.ID)
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
			return invalidRootStatusMessage(msg.ID)
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

func invalidRootStatusMessage(id string) serverMessage {
	return serverMessage{
		ID:      id,
		Kind:    "status",
		Status:  "invalid",
		Message: "invalid rootKey: not in predefined allowed set",
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

func (api *EventsAPI) emit(event eventEnvelope) {
	subscribers := api.matchingSubscribers(event.RootKey.KeyID)
	for _, sub := range subscribers {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = writeWS(ctx, sub.conn, serverMessage{
			Kind:  "event",
			Event: &event,
		})
		cancel()
	}
}

func (api *EventsAPI) matchingSubscribers(rootKeyID string) []*subscriber {
	api.mu.Lock()
	defer api.mu.Unlock()

	out := make([]*subscriber, 0, len(api.subscribers))
	for sub := range api.subscribers {
		sub.state.mu.Lock()
		_, ok := sub.state.roots[rootKeyID]
		sub.state.mu.Unlock()
		if ok {
			out = append(out, sub)
		}
	}
	return out
}

func (api *EventsAPI) addSubscriber(sub *subscriber) {
	api.mu.Lock()
	defer api.mu.Unlock()
	api.subscribers[sub] = struct{}{}
}

func (api *EventsAPI) removeSubscriber(sub *subscriber) {
	api.mu.Lock()
	defer api.mu.Unlock()
	delete(api.subscribers, sub)
}

func (api *EventsAPI) subscriberCount() int {
	api.mu.Lock()
	defer api.mu.Unlock()
	return len(api.subscribers)
}
