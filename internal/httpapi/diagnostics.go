package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
)

const defaultLogEntryLimit = 200

type LogStore struct {
	mu      sync.Mutex
	limit   int
	entries []string
}

func NewLogStore(limit int) *LogStore {
	if limit <= 0 {
		limit = defaultLogEntryLimit
	}
	return &LogStore{
		limit:   limit,
		entries: make([]string, 0, limit),
	}
}

func (s *LogStore) Add(entry string) {
	if s == nil {
		return
	}
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.entries) == s.limit {
		copy(s.entries, s.entries[1:])
		s.entries[len(s.entries)-1] = entry
		return
	}
	s.entries = append(s.entries, entry)
}

func (s *LogStore) Content() string {
	if s == nil {
		return ""
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return strings.Join(append([]string(nil), s.entries...), "\n")
}

func WithRequestLogging(next http.Handler, logs *LogStore) http.Handler {
	if logs == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusCapturingResponseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}
		next.ServeHTTP(rec, r)
		logs.Add("http method=" + r.Method + " path=" + r.URL.Path + " status=" + strconv.Itoa(rec.statusToWrite()))
	})
}

type statusCapturingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusCapturingResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusCapturingResponseWriter) statusToWrite() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}
