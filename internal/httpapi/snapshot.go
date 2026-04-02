package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/flarebyte/chatty-ratatoskr/internal/snapshot"
	"github.com/flarebyte/chatty-ratatoskr/internal/yggkey"
)

const defaultGeneratedResponseID = "generated"

type SnapshotAPI struct {
	store      snapshot.Store
	generateID func() string
}

type keyParams struct {
	KeyID       string `json:"keyId"`
	SecureKeyID string `json:"secureKeyId,omitempty"`
	Version     string `json:"version,omitempty"`
}

type keyValueParams struct {
	Key   keyParams `json:"key"`
	Value string    `json:"value,omitempty"`
}

type setSnapshotRequest struct {
	ID           string           `json:"id,omitempty"`
	Key          keyParams        `json:"key"`
	KeyValueList []keyValueParams `json:"keyValueList"`
}

type getSnapshotRequest struct {
	ID  string    `json:"id,omitempty"`
	Key keyParams `json:"key"`
}

type responseEnvelope[T any] struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data"`
}

type setSnapshotResponseData struct {
	Key keyParams `json:"key"`
}

type getSnapshotResponseData struct {
	Key          keyParams        `json:"key"`
	KeyValueList []keyValueParams `json:"keyValueList"`
}

func NewSnapshotAPI(store snapshot.Store) *SnapshotAPI {
	return &SnapshotAPI{
		store:      store,
		generateID: func() string { return defaultGeneratedResponseID },
	}
}

func NewSnapshotAPIWithGenerator(store snapshot.Store, generateID func() string) *SnapshotAPI {
	api := NewSnapshotAPI(store)
	if generateID != nil {
		api.generateID = generateID
	}
	return api
}

func (api *SnapshotAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("/snapshot", api.handleSnapshot)
}

func (api *SnapshotAPI) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.handleGetSnapshot(w, r)
	case http.MethodPut:
		api.handleSetSnapshot(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (api *SnapshotAPI) handleSetSnapshot(w http.ResponseWriter, r *http.Request) {
	var req setSnapshotRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.invalidEnvelope(req.ID, "invalid JSON payload"))
		return
	}

	rootParsed, err := yggkey.Parse(req.Key.KeyID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.invalidEnvelope(req.ID, err.Error()))
		return
	}

	entries := make([]snapshot.KeyValue, 0, len(req.KeyValueList))
	for _, item := range req.KeyValueList {
		if _, err := yggkey.Parse(item.Key.KeyID); err != nil {
			writeJSON(w, http.StatusBadRequest, api.invalidEnvelope(req.ID, err.Error()))
			return
		}
		if !isDescendant(rootParsed.Canonical, item.Key.KeyID) {
			writeJSON(w, http.StatusBadRequest, api.invalidEnvelope(req.ID, "invalid key: snapshot entry must be a descendant of the requested root"))
			return
		}
		entries = append(entries, snapshot.KeyValue{
			Key: snapshot.Key{
				KeyID:       item.Key.KeyID,
				SecureKeyID: item.Key.SecureKeyID,
				Version:     item.Key.Version,
			},
			Value: item.Value,
		})
	}

	root := snapshot.Key{
		KeyID:       req.Key.KeyID,
		SecureKeyID: req.Key.SecureKeyID,
		Version:     req.Key.Version,
	}
	api.store.Replace(root, entries)

	writeJSON(w, http.StatusOK, responseEnvelope[setSnapshotResponseData]{
		ID:     api.responseID(req.ID),
		Status: "ok",
		Data: setSnapshotResponseData{
			Key: req.Key,
		},
	})
}

func (api *SnapshotAPI) handleGetSnapshot(w http.ResponseWriter, r *http.Request) {
	var req getSnapshotRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.invalidEnvelope(req.ID, "invalid JSON payload"))
		return
	}

	if _, err := yggkey.Parse(req.Key.KeyID); err != nil {
		writeJSON(w, http.StatusBadRequest, api.invalidEnvelope(req.ID, err.Error()))
		return
	}

	root := snapshot.Key{
		KeyID:       req.Key.KeyID,
		SecureKeyID: req.Key.SecureKeyID,
		Version:     req.Key.Version,
	}
	got := api.store.Get(root)

	keyValueList := make([]keyValueParams, 0, len(got.KeyValueList))
	for _, item := range got.KeyValueList {
		keyValueList = append(keyValueList, keyValueParams{
			Key: keyParams{
				KeyID:       item.Key.KeyID,
				SecureKeyID: item.Key.SecureKeyID,
				Version:     item.Key.Version,
			},
			Value: item.Value,
		})
	}

	writeJSON(w, http.StatusOK, responseEnvelope[getSnapshotResponseData]{
		ID:     api.responseID(req.ID),
		Status: "ok",
		Data: getSnapshotResponseData{
			Key:          req.Key,
			KeyValueList: keyValueList,
		},
	})
}

func decodeJSON(r *http.Request, out any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (api *SnapshotAPI) responseID(requestID string) string {
	if requestID != "" {
		return requestID
	}
	return api.generateID()
}

func (api *SnapshotAPI) invalidEnvelope(requestID, message string) responseEnvelope[map[string]any] {
	return responseEnvelope[map[string]any]{
		ID:      api.responseID(requestID),
		Status:  "invalid",
		Message: message,
		Data:    map[string]any{},
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

func isDescendant(rootKey, childKey string) bool {
	return strings.HasPrefix(childKey, rootKey+":")
}
