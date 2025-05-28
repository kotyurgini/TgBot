package fileslog

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

type FileSlogHandler struct {
	mu    sync.Mutex
	path  string
	attrs []slog.Attr
}

func NewFileSlogHandler(path string) *FileSlogHandler {
	return &FileSlogHandler{
		path: path,
	}
}

func (h *FileSlogHandler) Handle(_ context.Context, r slog.Record) error {
	level := fmt.Sprintf("[%s]:", r.Level.String())

	fields := make(map[string]any, r.NumAttrs())

	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})

	for _, a := range h.attrs {
		fields[a.Key] = a.Value.Any()
	}

	var b []byte
	var err error

	if len(fields) > 0 {
		//b, err = json.MarshalIndent(fields, "", "  ")
		b, err = json.Marshal(fields)
		if err != nil {
			return err
		}
	}

	timeStr := r.Time.Format("[15:05:05.000]")

	if err := os.MkdirAll(h.path, 0755); err != nil {
		return err
	}

	today := strings.Replace(time.Now().UTC().Format("2006-01-02"), "-", "", -1)
	f, err := os.OpenFile(fmt.Sprintf("%s/%s.log", h.path, today), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	h.mu.Lock()
	defer h.mu.Unlock()

	str := fmt.Sprintf("%s %s %s %s\n", timeStr, level, r.Message, string(b))
	_, err = f.WriteString(str)
	return err
}

func (h *FileSlogHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *FileSlogHandler) WithGroup(_ string) slog.Handler {
	return h
}

func (h *FileSlogHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}
