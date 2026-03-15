package store

import (
	"log/slog"
	"sync"
	"time"
)

type Session struct {
	UserID    int64
	CreatedAt time.Time
	ExpiresAt time.Time
}

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]Session

	done chan struct{}
}

func NewSessionStore() *SessionStore {
	s := &SessionStore{
		sessions: make(map[string]Session),
		done:     make(chan struct{}),
	}

	//TODO(janitor): This should be configurable through settings
	go s.RunJanitor(5 * time.Minute)
	return s
}

func (s *SessionStore) Get(token string) (Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[token]
	return sess, ok
}

func (s *SessionStore) Put(token string, sess Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[token] = sess
}

func (s *SessionStore) Delete(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
}

func (s *SessionStore) Close() {
	close(s.done)
}

func (s *SessionStore) RunJanitor(interval time.Duration) {
	// Set interval < 0 to never clean up automatically
	if interval <= 0 {
		slog.Warn("Session storage will not be automatically pruned.")
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			now := time.Now()
			for t, sess := range s.sessions {
				if now.After(sess.ExpiresAt) {
					delete(s.sessions, t)
				}
			}
			s.mu.Unlock()
		case <-s.done:
			return
		}
	}
}
