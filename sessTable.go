// Copyright 2018 Cibifang. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package janus

import (
)

type sessTable struct {
    sessions    map[int]*Session
    msgs        map[string](chan ServerMsg)
}

func newSessTable() *sessTable {
    return &sessTable{sessions: make(map[int]*Session),
                      msgs: make(map[string](chan ServerMsg))}
}

func (st *sessTable) newSess(id int) *Session {
    st.sessions[id] = newSess(id)

    return st.sessions[id]
}

func (st *sessTable) newTransaction() string {
    tid := newTransaction()

    _, exist := st.msgs[tid]
    for exist {
        tid = newTransaction()
        _, exist = st.msgs[tid]
    }

    st.msgs[tid] = make(chan ServerMsg, 1)
    return tid
}

func (st *sessTable) MsgChan(tid string) (msgChan (chan ServerMsg), exist bool) {
    msgChan, exist = st.msgs[tid]
    return msgChan, exist
}
