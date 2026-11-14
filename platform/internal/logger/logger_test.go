package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestTraceHandler(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(traceHandler{slog.NewJSONHandler(&buf, nil)}).With("component", "test")

	traceID, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	spanID, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	}))

	t.Run("adds span IDs when the context has a span", func(t *testing.T) {
		buf.Reset()
		log.InfoContext(ctx, "hello")

		var record map[string]any
		require.NoError(t, json.Unmarshal(buf.Bytes(), &record))
		assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", record["trace_id"])
		assert.Equal(t, "00f067aa0ba902b7", record["span_id"])
		assert.Equal(t, "test", record["component"], "attrs from With are kept")
	})

	t.Run("adds nothing without a span", func(t *testing.T) {
		buf.Reset()
		log.InfoContext(context.Background(), "hello")

		var record map[string]any
		require.NoError(t, json.Unmarshal(buf.Bytes(), &record))
		assert.NotContains(t, record, "trace_id")
	})
}
