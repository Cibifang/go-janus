// Copyright 2018 Cibifang. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package janus

import (
    "sync"
)

type Handle struct {
    lock sync.Mutex
    id   uint64
    msgs map[string](chan []byte)
}

func newHandle(id uint64) *Handle {
    return &Handle{id: id,
        msgs: make(map[string](chan []byte))}
}

func (h *Handle) NewTransaction() string {
    tid := newTransaction()

    h.lock.Lock()
    defer h.lock.Unlock()
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

func (h *Handle) DefaultMsgChan() (chan []byte) {
    h.lock.Lock()
    defer h.lock.Unlock()
    _, exist := h.msgs[""]
    if !exist {
        h.msgs[""] = make(chan []byte, 1)
    }

    return h.msgs[""]
}
