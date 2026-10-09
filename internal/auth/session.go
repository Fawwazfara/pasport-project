package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const SessionTTL = 12 * time.Hour

type session struct {
	username string
	expires  time.Time
}

type SessionStore struct {
	mu       sync.Mutex
	sessions map[string]session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]session)}
}

func (s *SessionStore) Create(username string) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	id := hex.EncodeToString(buf)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = session{username: username, expires: time.Now().Add(SessionTTL)}
	return id, nil
}

func (s *SessionStore) Lookup(id string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[id]
	if !ok {
		return "", false
	}
	if time.Now().After(sess.expires) {
		delete(s.sessions, id)
		return "", false
	}
	return sess.username, true
}

func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}
