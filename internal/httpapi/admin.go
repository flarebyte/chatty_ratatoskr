package httpapi

import (
	"net/http"

	"github.com/flarebyte/chatty-ratatoskr/internal/snapshot"
)

type AdminAPI struct {
	store             snapshot.Store
	logs              *LogStore
	generateID        func() string
	payloadLimitBytes int64
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

type getCommandRequest struct {
	ID      string        `json:"id,omitempty"`
	Command commandParams `json:"command"`
}

type commandStatus struct {
	Command commandParams `json:"command"`
	Status  string        `json:"status"`
	Message string        `json:"message,omitempty"`
}

type setCommandsResponseData struct {
	Results []commandStatus `json:"results"`
}

type getCommandResponseData struct {
	Command commandParams `json:"command"`
	Content string        `json:"content"`
}

func NewAdminAPI(store snapshot.Store) *AdminAPI {
	return &AdminAPI{
		store:             store,
		generateID:        func() string { return defaultGeneratedResponseID },
		payloadLimitBytes: defaultHTTPPayloadLimitBytes,
	}
}

func NewAdminAPIWithLimit(store snapshot.Store, payloadLimitBytes int64) *AdminAPI {
	api := NewAdminAPI(store)
	if payloadLimitBytes > 0 {
		api.payloadLimitBytes = payloadLimitBytes
	}
	return api
}

func NewAdminAPIWithOptions(store snapshot.Store, logs *LogStore, payloadLimitBytes int64) *AdminAPI {
	api := NewAdminAPIWithLimit(store, payloadLimitBytes)
	api.logs = logs
	return api
}

func (api *AdminAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("/admin/commands", api.handleCommands)
}

func (api *AdminAPI) handleCommands(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.handleGetCommand(w, r)
	case http.MethodPut:
		api.handleSetCommands(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (api *AdminAPI) handleGetCommand(w http.ResponseWriter, r *http.Request) {
	var req getCommandRequest
	if err := decodeJSONWithLimit(r, &req, api.payloadLimitBytes); err != nil {
		writeJSON(w, statusForDecodeError(err), invalidEnvelopeWithID(req.ID, api.generateID, messageForDecodeError(err)))
		return
	}

	switch req.Command.ID {
	case "read-logs":
		writeJSON(w, http.StatusOK, responseEnvelope[getCommandResponseData]{
			ID:     responseIDWithGenerator(req.ID, api.generateID),
			Status: "ok",
			Data: getCommandResponseData{
				Command: req.Command,
				Content: api.logs.Content(),
			},
		})
	default:
		writeJSON(w, http.StatusBadRequest, invalidEnvelopeWithID(req.ID, api.generateID, "unknown admin command"))
	}
}

func (api *AdminAPI) handleSetCommands(w http.ResponseWriter, r *http.Request) {
	var req setCommandsRequest
	if err := decodeJSONWithLimit(r, &req, api.payloadLimitBytes); err != nil {
		writeJSON(w, statusForDecodeError(err), invalidEnvelopeWithID(req.ID, api.generateID, messageForDecodeError(err)))
		return
	}

	results := make([]commandStatus, 0, len(req.Commands))
	for _, cmd := range req.Commands {
		switch cmd.ID {
		case "clear-state":
			api.store.Clear()
			api.logs.Add("admin command=clear-state status=ok")
			results = append(results, commandStatus{
				Command: cmd,
				Status:  "ok",
			})
		default:
			writeJSON(w, http.StatusBadRequest, invalidEnvelopeWithID(req.ID, api.generateID, "unknown admin command"))
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
