// Copyright 2018 Cibifang. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package janus

import ()

type Handle struct {
    id   uint64
    msgs map[string](chan []byte)
}

func newHandle(id uint64) *Handle {
    return &Handle{id: id,
        msgs: make(map[string](chan []byte))}
}

func (h *Handle) NewTransaction() string {
    tid := newTransaction()

    _, exist := h.msgs[tid]
    for exist {
        tid = newTransaction()
        _, exist = h.msgs[tid]
    }

    h.msgs[tid] = make(chan []byte, 1)
    return tid
}

func (h *Handle) MsgChan(tid string) (msgChan chan []byte, exist bool) {
    msgChan, exist = h.msgs[tid]
    return msgChan, exist
}
