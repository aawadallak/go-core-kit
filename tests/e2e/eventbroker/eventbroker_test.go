package eventbroker_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aawadallak/go-core-kit/core/event"
	"github.com/aawadallak/go-core-kit/plugin/event/eventbroker"
)

// memTransport is an in-memory eventbroker.Transport used for testing.
type memTransport struct {
	mu       sync.Mutex
	messages []struct {
		Subject string
		Data    []byte
	}
}

func (m *memTransport) Publish(_ context.Context, subject string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, struct {
		Subject string
		Data    []byte
	}{Subject: subject, Data: data})
	return nil
}

func (m *memTransport) get(i int) (subject string, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.messages[i].Subject, m.messages[i].Data
}

func (m *memTransport) len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.messages)
}

// testMetadata implements event.Metadata for test purposes.
type testMetadata struct {
	UserID string `json:"user_id"`
}

func (t testMetadata) EventType() string { return "USER_CREATED" }
func (t testMetadata) EventVersion() int { return 1 }

func TestDispatcher_Dispatch(t *testing.T) {
	transport := &memTransport{}
	subject := "events.user"
	dispatcher := eventbroker.NewDispatcher(transport, subject)

	record, err := event.NewRecord(testMetadata{UserID: "u-123"}, event.WithCorrelationID("corr-1"))
	require.NoError(t, err)

	err = dispatcher.Dispatch(context.Background(), record)
	require.NoError(t, err)

	require.Equal(t, 1, transport.len())

	gotSubject, data := transport.get(0)
	assert.Equal(t, subject, gotSubject)

	var env eventbroker.Envelope
	require.NoError(t, json.Unmarshal(data, &env))

	assert.Equal(t, "USER_CREATED", env.EventName)
	assert.Equal(t, 1, env.Version)
	assert.Equal(t, "corr-1", env.CorrelationID)
	assert.Equal(t, record.ID, env.EventID)
	assert.False(t, env.OccurredAt.IsZero())

	// Verify the payload round-trips.
	var meta testMetadata
	require.NoError(t, json.Unmarshal(env.Payload, &meta))
	assert.Equal(t, "u-123", meta.UserID)
}

func TestDispatcher_MultipleEvents(t *testing.T) {
	transport := &memTransport{}
	dispatcher := eventbroker.NewDispatcher(transport, "events.multi")

	for range 5 {
		record, err := event.NewRecord(testMetadata{UserID: "u-multi"})
		require.NoError(t, err)
		require.NoError(t, dispatcher.Dispatch(context.Background(), record))
	}

	assert.Equal(t, 5, transport.len())
}

func TestDispatcher_PreservesTimestamp(t *testing.T) {
	transport := &memTransport{}
	dispatcher := eventbroker.NewDispatcher(transport, "events.ts")

	before := time.Now().UTC().Add(-time.Second)

	record, err := event.NewRecord(testMetadata{UserID: "ts"})
	require.NoError(t, err)
	require.NoError(t, dispatcher.Dispatch(context.Background(), record))

	_, data := transport.get(0)
	var env eventbroker.Envelope
	require.NoError(t, json.Unmarshal(data, &env))

	assert.True(t, env.OccurredAt.After(before), "OccurredAt should be after test start")
}
