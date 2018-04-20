// Copyright 2018 Cibifang. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package janus

import (
)

type Session struct {
    id          int
    handles     map[int]*Handle
    msgs        map[string](chan ServerMsg)
}

func newSess(id int) *Session {
    return &Session{id: id,
                    handles: make(map[int]*Handle),
                    msgs: make(map[string](chan ServerMsg)),}
}


/* because now we just use one plugin, so don't create plugin.go */
func (s *Session) Attach(id int) (*Handle) {
    s.handles[id] = newHandle(id)
    return s.handles[id]
}

func (s *Session) NewTransaction() string {
    tid := newTransaction()

    _, exist := s.msgs[tid]
    for exist {
        tid = newTransaction()
        _, exist = s.msgs[tid]
    }

    s.msgs[tid] = make(chan ServerMsg, 1)
    return tid
}

func (s *Session) MsgChan(tid string) (msgChan (chan ServerMsg), exist bool) {
    msgChan, exist = s.msgs[tid]
    return msgChan, exist
}
