package httpapi

import (
	"net/http"

	"github.com/flarebyte/chatty-ratatoskr/internal/snapshot"
	"github.com/flarebyte/chatty-ratatoskr/internal/yggkey"
)

type NodeAPI struct {
	store      snapshot.Store
	generateID func() string
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

		api.store.Upsert(root, snapshot.KeyValue{
			Key: snapshot.Key{
				KeyID:       item.Key.KeyID,
				SecureKeyID: item.Key.SecureKeyID,
				Version:     item.Key.Version,
			},
			Value: item.Value,
		})
		keyList = append(keyList, result)
	}

	writeJSON(w, http.StatusOK, responseEnvelope[setKeyValueResponseData]{
		ID:     responseIDWithGenerator(req.ID, api.generateID),
		Status: "ok",
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
