// Copyright 2018 Cibifang. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package janus

import (
	"sync"
)

type Session struct {
	lock    sync.Mutex
	id      uint64
	handles map[uint64]*Handle
	msgs    map[string](chan []byte)
}

func newSess(id uint64) *Session {
	return &Session{
		id:      id,
		handles: make(map[uint64]*Handle),
		msgs:    make(map[string](chan []byte))}
}

/* because now we just use one plugin, so don't create plugin.go */
func (s *Session) Attach(id uint64) *Handle {
	s.handles[id] = newHandle(id)
	return s.handles[id]
}

func (s *Session) NewTransaction() string {
	tid := newTransaction()

	s.lock.Lock()
	defer s.lock.Unlock()
	_, exist := s.msgs[tid]
	for exist {
		tid = newTransaction()
		_, exist = s.msgs[tid]
	}

	s.msgs[tid] = make(chan []byte, 1)
	return tid
}

func (s *Session) MsgChan(tid string) (msgChan chan []byte, exist bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	msgChan, exist = s.msgs[tid]
	return msgChan, exist
}

func (s *Session) DefaultMsgChan() chan []byte {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, exist := s.msgs[""]
	if !exist {
		s.msgs[""] = make(chan []byte, 1)
	}

	return s.msgs[""]
}
