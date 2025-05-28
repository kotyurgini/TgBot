package myslog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	stdLog "log"
	"log/slog"
)

type MyHandlerOptions struct {
	SlogOpts *slog.HandlerOptions
}

type MyHandler struct {
	slog.Handler
	l     *stdLog.Logger
	attrs []slog.Attr
}

func (opts MyHandlerOptions) NewMyHandler(out io.Writer) *MyHandler {
	return &MyHandler{
		Handler: slog.NewTextHandler(out, opts.SlogOpts),
		l:       stdLog.New(out, "", 0),
	}
}

func (h *MyHandler) Handle(_ context.Context, r slog.Record) error {
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

	timeStr := r.Time.UTC().Format("[2006-01-02 15:05:05.000]")

	h.l.Println(timeStr, level, r.Message, string(b))

	return nil
}

func (h *MyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &MyHandler{
		Handler: h.Handler,
		l:       h.l,
		attrs:   attrs,
	}
}
