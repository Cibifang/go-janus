// Copyright 2018 Cibifang. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package janus

import (
    "time"
    "math/rand"
    "log"
)


type ClientBody struct {
    Request     string `json:"request,omitempty"`
}

// WebSocket message send to janus.
type ClientMsg struct {
    Janus           string `json:"janus,omitempty"`
    Plugin          string `json:"plugin,omitempty"`
    Transaction     string `json:"transaction,omitempty"`
    SessionId       int `json:"session_id,omitempty"`
    HandleId        int `json:"handle_id,omitempty"`
    Body            ClientBody `json:"body,omitempty"`
}

type ServerData struct {
    Id              int `json:"id,omitempty"`
}

type ServerMsg struct {
    Janus           string `json:"janus,omitempty"`
    Transaction     string `json:"transaction,omitempty"`
    Data            ServerData `json:"data,omitempty"`
    SessionId       int `json:"session_id,omitempty"`
    HandleId        int `json:"handle_id,omitempty"`
}

type ServerMsgChan interface {
    MsgChan(string) ((chan ServerMsg), bool)
}

const transactionSize = 12

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
    letterIdxBits = 6                    // 6 bits to represent a letter index
    letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
    letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)


func randString(n int) string {
    var randSeed = rand.NewSource(time.Now().UnixNano())
    b := make([]byte, n)
    // A randSeed.Int63() generates 63 random bits,
    // enough for letterIdxMax characters!
    for i, cache, remain := n-1, randSeed.Int63(), letterIdxMax; i >= 0; {
        if remain == 0 {
            cache, remain = randSeed.Int63(), letterIdxMax
        }
        if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
            b[i] = letterBytes[idx]
            i--
        }
        cache >>= letterIdxBits
        remain--
    }

    return string(b)
}

func newTransaction() string {
    return randString(transactionSize)
}

func transferServerMsg(tid string, msg ServerMsg, serverMsgChan ServerMsgChan) {
    if msgChan, ok := serverMsgChan.MsgChan(tid); ok {
        log.Printf("Transfer message %+v to channel with tid %s", msg, tid)
        msgChan <- msg
    } else {
        log.Printf("main: can't find channel with tid %s", tid)
    }
}
