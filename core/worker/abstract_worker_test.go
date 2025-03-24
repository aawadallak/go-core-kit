package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockProvider struct {
	interval      time.Duration
	executeErr    error
	onStartErr    error
	onPreExecErr  error
	onPostExecErr error
	onShutdownErr error
}

func (m *mockProvider) Interval() time.Duration {
	return m.interval
}

func (m *mockProvider) Execute(ctx context.Context) error {
	return m.executeErr
}

func (m *mockProvider) OnStart(ctx context.Context) error {
	return m.onStartErr
}

func (m *mockProvider) OnPreExecute(ctx context.Context) error {
	return m.onPreExecErr
}

func (m *mockProvider) OnPostExecute(ctx context.Context) error {
	return m.onPostExecErr
}

func (m *mockProvider) OnShutdown(ctx context.Context) error {
	return m.onShutdownErr
}

func TestWorker(t *testing.T) {
	tests := []struct {
		name          string
		provider      *mockProvider
		options       []Option
		expectedError error
	}{
		{
			name: "successful execution",
			provider: &mockProvider{
				interval: time.Millisecond * 10,
			},
			expectedError: nil,
		},
		{
			name: "execution with start error",
			provider: &mockProvider{
				interval:   time.Millisecond * 10,
				onStartErr: errors.New("start error"),
			},
			expectedError: errors.New("start error"),
		},
		{
			name: "execution with pre-execute error",
			provider: &mockProvider{
				interval:     time.Millisecond * 10,
				onPreExecErr: errors.New("pre-execute error"),
			},
			expectedError: nil,
		},
		{
			name: "execution with execute error",
			provider: &mockProvider{
				interval:   time.Millisecond * 10,
				executeErr: errors.New("execute error"),
			},
			expectedError: nil,
		},
		{
			name: "execution with post-execute error",
			provider: &mockProvider{
				interval:      time.Millisecond * 10,
				onPostExecErr: errors.New("post-execute error"),
			},
			expectedError: nil,
		},
		{
			name: "execution with abort on error",
			provider: &mockProvider{
				interval:   time.Millisecond * 10,
				executeErr: errors.New("execute error"),
			},
			options:       []Option{WithAbortOnError()},
			expectedError: nil,
		},
		{
			name: "synchronous execution",
			provider: &mockProvider{
				interval: time.Millisecond * 10,
			},
			options:       []Option{WithSynchronousExecution()},
			expectedError: nil,
		},
		{
			name: "abort on error stops execution",
			provider: &mockProvider{
				interval:   time.Millisecond * 10,
				executeErr: errors.New("execute error"),
				onStartErr: nil,
			},
			options:       []Option{WithAbortOnError(), WithSynchronousExecution()},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			worker := NewAbstractWorker(tt.provider, tt.options...)
			err := worker.Start(ctx)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
