package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/flarebyte/chatty-ratatoskr/internal/snapshot"
	"github.com/flarebyte/chatty-ratatoskr/internal/yggkey"
)

type NodeAPI struct {
	store      snapshot.Store
	generateID func() string
	events     *EventsAPI
}

type setKeyValueRequest struct {
	ID           string           `json:"id,omitempty"`
	RootKey      keyParams        `json:"rootKey"`
	KeyValueList []keyValueParams `json:"keyValueList"`
}

type getKeyValueRequest struct {
	ID      string      `json:"id,omitempty"`
	RootKey keyParams   `json:"rootKey"`
	KeyList []keyParams `json:"keyList"`
}

type keyStatusResult struct {
	Key     keyParams `json:"key"`
	Status  string    `json:"status"`
	Message string    `json:"message,omitempty"`
}

type setKeyValueResponseData struct {
	RootKey keyParams         `json:"rootKey"`
	KeyList []keyStatusResult `json:"keyList"`
}

type keyValueStatusResult struct {
	KeyValue keyValueParams `json:"keyValue"`
	Status   string         `json:"status"`
	Message  string         `json:"message,omitempty"`
}

type getKeyValueResponseData struct {
	RootKey      keyParams              `json:"rootKey"`
	KeyValueList []keyValueStatusResult `json:"keyValueList"`
}

func NewNodeAPI(store snapshot.Store) *NodeAPI {
	return &NodeAPI{
		store:      store,
		generateID: func() string { return defaultGeneratedResponseID },
	}
}

func NewNodeAPIWithEvents(store snapshot.Store, events *EventsAPI) *NodeAPI {
	api := NewNodeAPI(store)
	api.events = events
	return api
}

func (api *NodeAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("/node", api.handleNode)
}

func (api *NodeAPI) handleNode(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.handleGetKeyValueList(w, r)
	case http.MethodPut:
		api.handleSetKeyValueList(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (api *NodeAPI) handleSetKeyValueList(w http.ResponseWriter, r *http.Request) {
	var req setKeyValueRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, responseEnvelope[map[string]any]{
			ID:      responseIDWithGenerator(req.ID, api.generateID),
			Status:  "invalid",
			Message: "invalid JSON payload",
			Data:    map[string]any{},
		})
		return
	}

	rootParsed, err := yggkey.Parse(req.RootKey.KeyID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, responseEnvelope[map[string]any]{
			ID:      responseIDWithGenerator(req.ID, api.generateID),
			Status:  "invalid",
			Message: err.Error(),
			Data:    map[string]any{},
		})
		return
	}

	root := snapshot.Key{
		KeyID:       req.RootKey.KeyID,
		SecureKeyID: req.RootKey.SecureKeyID,
		Version:     req.RootKey.Version,
	}

	keyList := make([]keyStatusResult, 0, len(req.KeyValueList))
	hasOutdated := false
	for _, item := range req.KeyValueList {
		result := keyStatusResult{
			Key: keyParams{
				KeyID:       item.Key.KeyID,
				SecureKeyID: item.Key.SecureKeyID,
				Version:     item.Key.Version,
			},
			Status: "ok",
		}

		parsed, parseErr := yggkey.Parse(item.Key.KeyID)
		if parseErr != nil {
			result.Status = "invalid"
			result.Message = parseErr.Error()
			keyList = append(keyList, result)
			continue
		}

		result.Key.Kind = derivedKindParams(parsed)

		if !isDescendant(rootParsed.Canonical, item.Key.KeyID) {
			result.Status = "invalid"
			result.Message = "invalid key: node entry must be a descendant of the requested root"
			keyList = append(keyList, result)
			continue
		}

		existing, exists := api.store.Find(root, item.Key.KeyID)
		nextVersion := "v1"
		if exists {
			if item.Key.Version != existing.Key.Version {
				result.Status = "outdated"
				result.Message = "outdated version: write is not based on the latest stored version"
				hasOutdated = true
				keyList = append(keyList, result)
				continue
			}
			nextVersion = bumpVersion(existing.Key.Version)
		}

		result.Key.Version = nextVersion
		api.store.Upsert(root, snapshot.KeyValue{
			Key: snapshot.Key{
				KeyID:       item.Key.KeyID,
				SecureKeyID: item.Key.SecureKeyID,
				Version:     nextVersion,
			},
			Value: item.Value,
		})
		if api.events != nil {
			api.events.EmitSet(
				keyParams{
					KeyID:       req.RootKey.KeyID,
					SecureKeyID: req.RootKey.SecureKeyID,
					Version:     req.RootKey.Version,
					Kind:        derivedKindParams(rootParsed),
				},
				keyValueParams{
					Key:   result.Key,
					Value: item.Value,
				},
			)
		}
		keyList = append(keyList, result)
	}

	statusCode := http.StatusOK
	statusText := "ok"
	if hasOutdated {
		statusCode = http.StatusConflict
		statusText = "outdated"
	}

	writeJSON(w, statusCode, responseEnvelope[setKeyValueResponseData]{
		ID:     responseIDWithGenerator(req.ID, api.generateID),
		Status: statusText,
		Data: setKeyValueResponseData{
			RootKey: keyParams{
				KeyID:       req.RootKey.KeyID,
				SecureKeyID: req.RootKey.SecureKeyID,
				Version:     req.RootKey.Version,
				Kind:        derivedKindParams(rootParsed),
			},
			KeyList: keyList,
		},
	})
}

func bumpVersion(current string) string {
	if !strings.HasPrefix(current, "v") {
		if current == "" {
			return "v1"
		}
		return current + ".next"
	}
	number, err := strconv.Atoi(strings.TrimPrefix(current, "v"))
	if err != nil {
		return current + ".next"
	}
	return "v" + strconv.Itoa(number+1)
}

func (api *NodeAPI) handleGetKeyValueList(w http.ResponseWriter, r *http.Request) {
	var req getKeyValueRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, responseEnvelope[map[string]any]{
			ID:      responseIDWithGenerator(req.ID, api.generateID),
			Status:  "invalid",
			Message: "invalid JSON payload",
			Data:    map[string]any{},
		})
		return
	}

	rootParsed, err := yggkey.Parse(req.RootKey.KeyID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, responseEnvelope[map[string]any]{
			ID:      responseIDWithGenerator(req.ID, api.generateID),
			Status:  "invalid",
			Message: err.Error(),
			Data:    map[string]any{},
		})
		return
	}

	root := snapshot.Key{
		KeyID:       req.RootKey.KeyID,
		SecureKeyID: req.RootKey.SecureKeyID,
		Version:     req.RootKey.Version,
	}

	keyValueList := make([]keyValueStatusResult, 0, len(req.KeyList))
	for _, key := range req.KeyList {
		result := keyValueStatusResult{
			KeyValue: keyValueParams{
				Key: keyParams{
					KeyID:       key.KeyID,
					SecureKeyID: key.SecureKeyID,
					Version:     key.Version,
				},
			},
			Status: "ok",
		}

		parsed, parseErr := yggkey.Parse(key.KeyID)
		if parseErr != nil {
			result.Status = "invalid"
			result.Message = parseErr.Error()
			keyValueList = append(keyValueList, result)
			continue
		}

		result.KeyValue.Key.Kind = derivedKindParams(parsed)

		if !isDescendant(rootParsed.Canonical, key.KeyID) {
			result.Status = "invalid"
			result.Message = "invalid key: requested node must be a descendant of the requested root"
			keyValueList = append(keyValueList, result)
			continue
		}

		stored, ok := api.store.Find(root, key.KeyID)
		if !ok {
			result.Status = "invalid"
			result.Message = "missing key: requested node is not present under the requested root"
			keyValueList = append(keyValueList, result)
			continue
		}

		result.KeyValue.Key.SecureKeyID = stored.Key.SecureKeyID
		result.KeyValue.Key.Version = stored.Key.Version
		result.KeyValue.Value = stored.Value
		keyValueList = append(keyValueList, result)
	}

	writeJSON(w, http.StatusOK, responseEnvelope[getKeyValueResponseData]{
		ID:     responseIDWithGenerator(req.ID, api.generateID),
		Status: "ok",
		Data: getKeyValueResponseData{
			RootKey: keyParams{
				KeyID:       req.RootKey.KeyID,
				SecureKeyID: req.RootKey.SecureKeyID,
				Version:     req.RootKey.Version,
				Kind:        derivedKindParams(rootParsed),
			},
			KeyValueList: keyValueList,
		},
	})
}
