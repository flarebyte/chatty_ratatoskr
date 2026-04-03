package httpapi

import (
	"net/http"

	"github.com/flarebyte/chatty-ratatoskr/internal/snapshot"
)

type AdminAPI struct {
	store      snapshot.Store
	generateID func() string
}

type commandParams struct {
	ID        string   `json:"id"`
	Comment   string   `json:"comment,omitempty"`
	Arguments []string `json:"arguments,omitempty"`
}

type setCommandsRequest struct {
	ID       string          `json:"id,omitempty"`
	Commands []commandParams `json:"commands"`
}

type commandStatus struct {
	Command commandParams `json:"command"`
	Status  string        `json:"status"`
	Message string        `json:"message,omitempty"`
}

type setCommandsResponseData struct {
	Results []commandStatus `json:"results"`
}

func NewAdminAPI(store snapshot.Store) *AdminAPI {
	return &AdminAPI{
		store:      store,
		generateID: func() string { return defaultGeneratedResponseID },
	}
}

func (api *AdminAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("/admin/commands", api.handleCommands)
}

func (api *AdminAPI) handleCommands(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		api.handleSetCommands(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (api *AdminAPI) handleSetCommands(w http.ResponseWriter, r *http.Request) {
	var req setCommandsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, responseEnvelope[map[string]any]{
			ID:      responseIDWithGenerator(req.ID, api.generateID),
			Status:  "invalid",
			Message: "invalid JSON payload",
			Data:    map[string]any{},
		})
		return
	}

	results := make([]commandStatus, 0, len(req.Commands))
	for _, cmd := range req.Commands {
		switch cmd.ID {
		case "clear-state":
			api.store.Clear()
			results = append(results, commandStatus{
				Command: cmd,
				Status:  "ok",
			})
		default:
			writeJSON(w, http.StatusBadRequest, responseEnvelope[map[string]any]{
				ID:      responseIDWithGenerator(req.ID, api.generateID),
				Status:  "invalid",
				Message: "unknown admin command",
				Data:    map[string]any{},
			})
			return
		}
	}

	writeJSON(w, http.StatusOK, responseEnvelope[setCommandsResponseData]{
		ID:     responseIDWithGenerator(req.ID, api.generateID),
		Status: "ok",
		Data: setCommandsResponseData{
			Results: results,
		},
	})
}
