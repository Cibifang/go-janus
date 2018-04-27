// Copyright 2018 Cibifang. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package janus

import ()

type sessTable struct {
    sessions map[uint64]*Session
    msgs     map[string](chan []byte)
}

func newSessTable() *sessTable {
    return &sessTable{sessions: make(map[uint64]*Session),
                      msgs: make(map[string](chan []byte))}
}

func (st *sessTable) newSess(id uint64) *Session {
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

    st.msgs[tid] = make(chan []byte, 1)
    return tid
}

func (st *sessTable) MsgChan(tid string) (msgChan chan []byte, exist bool) {
    msgChan, exist = st.msgs[tid]
    return msgChan, exist
}
