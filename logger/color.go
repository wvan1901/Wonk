package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
)

const (
	ANSI_RESET_COLOR  = "\033[0m"
	ANSI_BLACK        = 30
	ANSI_RED          = 31
	ANSI_GREEN        = 32
	ANSI_YELLOW       = 33
	ANSI_BLUE         = 34
	ANSI_MAGENTA      = 35
	ANSI_CYAN         = 36
	ANSI_LIGHTGRAY    = 37
	ANSI_DARKGRAY     = 90
	ANSI_LIGHTRED     = 91
	ANSI_LIGHTGREEN   = 92
	ANSI_LIGHTYELLOW  = 93
	ANSI_LIGHTBLUE    = 94
	ANSI_LIGHTMAGENTA = 95
	ANSI_LIGHTCYAN    = 96
	ANSI_WHITE        = 97
	TIME_FORMAT       = "[15:04:05.000]"
)

func colorString(colorAnsiCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorAnsiCode), v, ANSI_RESET_COLOR)
}

type colorHandler struct {
	h slog.Handler  // Nested Slog handler
	b *bytes.Buffer // Capture the output from the "nested" handler
	m *sync.Mutex   // Thread safty to buf
}

func (h *colorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.h.Enabled(ctx, level)
}

func (h *colorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &colorHandler{h: h.h.WithAttrs(attrs), b: h.b, m: h.m}
}

func (h *colorHandler) WithGroup(name string) slog.Handler {
	return &colorHandler{h: h.h.WithGroup(name), b: h.b, m: h.m}
}

func (h *colorHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = colorString(ANSI_DARKGRAY, level)
	case slog.LevelInfo:
		level = colorString(ANSI_CYAN, level)
	case slog.LevelWarn:
		level = colorString(ANSI_LIGHTYELLOW, level)
	case slog.LevelError:
		level = colorString(ANSI_LIGHTRED, level)
	}

	attrs, err := h.computeAttrs(ctx, r)
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(attrs, "", "  ")
	if err != nil {
		return fmt.Errorf("error when marshaling attrs: %w", err)
	}
	fmt.Println(
		colorString(ANSI_LIGHTGRAY, r.Time.Format(TIME_FORMAT)),
		level,
		colorString(ANSI_WHITE, r.Message),
		colorString(ANSI_DARKGRAY, string(bytes)),
	)

	return nil
}

func (h *colorHandler) computeAttrs(
	ctx context.Context,
	r slog.Record,
) (map[string]any, error) {
	h.m.Lock()
	defer func() {
		h.b.Reset()
		h.m.Unlock()
	}()
	if err := h.h.Handle(ctx, r); err != nil {
		return nil, fmt.Errorf("error when calling inner handler's Handle: %w", err)
	}

	var attrs map[string]any
	err := json.Unmarshal(h.b.Bytes(), &attrs)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshaling inner handler's Handle result: %w", err)
	}
	return attrs, nil
}

func suppressDefaults(
	next func([]string, slog.Attr) slog.Attr,
) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey ||
			a.Key == slog.LevelKey ||
			a.Key == slog.MessageKey {
			return slog.Attr{}
		}
		if next == nil {
			return a
		}
		return next(groups, a)
	}
}

func newColorHandler(opts *slog.HandlerOptions) *colorHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	b := &bytes.Buffer{}
	return &colorHandler{
		b: b,
		h: slog.NewJSONHandler(b, &slog.HandlerOptions{
			Level:       opts.Level,
			AddSource:   opts.AddSource,
			ReplaceAttr: suppressDefaults(opts.ReplaceAttr),
		}),
		m: &sync.Mutex{},
	}
}
