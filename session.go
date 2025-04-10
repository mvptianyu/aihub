/*
@Project: aihub
@Module: aihub
@File : session.go
*/
package aihub

import (
	uuid "github.com/satori/go.uuid"
	"sync"
)

type Session struct {
	SessionID   string                 `json:"session_id"`
	SessionData map[string]interface{} `json:"session_data"`

	lock sync.RWMutex
}

func newSession(src map[string]interface{}) *Session {
	ret := &Session{
		SessionID:   uuid.NewV4().String(),
		SessionData: make(map[string]interface{}),
	}
	if src != nil {
		ret.SessionData = src
	}
	return ret
}

func (s *Session) SetSessionData(key string, value interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.SessionData == nil {
		s.SessionData = make(map[string]interface{})
	}
	s.SessionData[key] = value
}

func (s *Session) GetSessionData(key string) interface{} {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if tmp, ok := s.SessionData[key]; ok {
		return tmp
	}
	return nil
}

func (s *Session) GetAllSessionData() map[string]interface{} {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.SessionData
}

func (s *Session) GetSessionID() string {
	return s.SessionID
}

func (s *Session) MergeSessionData(data map[string]interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.SessionData == nil {
		s.SessionData = make(map[string]interface{})
	}
	for key, value := range data {
		s.SessionData[key] = value
	}
}
