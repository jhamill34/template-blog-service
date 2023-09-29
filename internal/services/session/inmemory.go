package session

import (
	"context"
	"fmt"

	"github.com/jhamill34/notion-provisioner/internal"
)

type InMemorySessionStore struct {
	sessions map[string]interface{}
}

// Create implements services.SessionService.
func (s *InMemorySessionStore) Create(_ context.Context, data interface{}) (string, error) {
	id := internal.GenerateId(256)
	s.sessions[id] = data
	return id, nil
}

// Destroy implements services.SessionService.
func (s *InMemorySessionStore) Destroy(_ context.Context, id string) error {
	delete(s.sessions, id)

	return nil
}

// Find implements services.SessionService.
func (s *InMemorySessionStore) Find(_ context.Context, id string, data interface{}) error {
	if found, ok := s.sessions[id]; ok {
		data = found
		return nil
	}

	return fmt.Errorf("Session with id %s not found", id)
}

func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		sessions: make(map[string]interface{}),
	}
}

// var _ services.SessionService = (*InMemorySessionStore)(nil)
