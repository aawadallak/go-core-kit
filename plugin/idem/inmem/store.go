package inmem

import (
	"context"
	"sync"
	"time"

	"github.com/aawadallak/go-core-kit/core/idem"
)

type Store struct {
	mu      sync.Mutex
	records map[string]idem.Record
}

func NewStore() *Store {
	return &Store{records: make(map[string]idem.Record)}
}

func (s *Store) Claim(_ context.Context, key string, opts idem.ClaimOptions) (idem.ClaimResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if rec, ok := s.records[key]; ok {
		switch rec.Status {
		case idem.StatusCompleted, idem.StatusFailed, idem.StatusDropped:
			return idem.ClaimResult{Acquired: false, Record: rec}, nil
		case idem.StatusProcessing:
			if rec.LeaseUntil != nil && rec.LeaseUntil.After(now) {
				return idem.ClaimResult{Acquired: false, Record: rec}, nil
			}
		}
	}

	lease := now.Add(opts.TTL)
	rec := idem.Record{
		Key:        key,
		Status:     idem.StatusProcessing,
		Owner:      opts.Owner,
		LeaseUntil: &lease,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if prev, ok := s.records[key]; ok {
		rec.CreatedAt = prev.CreatedAt
	}

	s.records[key] = rec
	return idem.ClaimResult{Acquired: true, Record: rec}, nil
}

func (s *Store) Complete(_ context.Context, key string, outcome []byte) (idem.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	rec := s.records[key]
	rec.Status = idem.StatusCompleted
	rec.Outcome = append([]byte(nil), outcome...)
	rec.UpdatedAt = now
	rec.CompletedAt = &now
	rec.LeaseUntil = nil
	s.records[key] = rec
	return rec, nil
}

func (s *Store) Fail(_ context.Context, key string, outcome []byte, status idem.Status) (idem.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	rec := s.records[key]
	rec.Status = status
	rec.Outcome = append([]byte(nil), outcome...)
	rec.UpdatedAt = now
	rec.CompletedAt = &now
	rec.LeaseUntil = nil
	s.records[key] = rec
	return rec, nil
}

func (s *Store) Get(_ context.Context, key string) (idem.Record, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rec, ok := s.records[key]
	return rec, ok, nil
}
