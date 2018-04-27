// Copyright 2018 Cibifang. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package janus

import (
    "log"
    "net/http"
    "time"

    "github.com/gorilla/websocket"
    "github.com/tidwall/gjson"
)

type Janus struct {
    *sessTable
    remoteServer string
    connected    bool
    sendChan     chan interface{}
}

var janusDial = websocket.Dialer{
    Proxy:            http.ProxyFromEnvironment,
    HandshakeTimeout: 45 * time.Second,
    Subprotocols:     []string{"janus-protocol"},
}

func NewJanus(rs string) *Janus {
    return &Janus{remoteServer: rs,
                  sessTable: newSessTable(),
                  connected: false,
                  sendChan:  make(chan interface{})}
}

func (j *Janus) WaitMsg() {
    conn, _, err := janusDial.Dial(j.remoteServer, nil)
    if err != nil {
        return
    }

    j.connected = true
    defer func() {
        j.connected = false
    }()

    closeChan := make(chan struct{})
    read := make(chan []byte)

    go func() {
        defer func() {
            closeChan <- struct{}{}
        }()
        for {
            _, msg, err := conn.ReadMessage();
            if err != nil {
                log.Println("c.ReadJSON error: ", err)
                break
            }
            read <- msg
        }
    }()

    var msg []byte
    var sendMsg interface{}

loop:
    for {
        select {
        case <-closeChan:
            break loop
        case msg = <-read:
            log.Printf("janus: receive message %s", msg)
            tid := gjson.GetBytes(msg, "transaction").String()
            sid := gjson.GetBytes(msg, "session_id")
            if !sid.Exists() {
                transferServerMsg(tid, msg, j)
                break
            }

            sess, ok := j.Session(sid.Uint())
            if !ok {
                log.Printf("janus: can't find session with id %d", sid.Uint())
                break
            }

            hid := gjson.GetBytes(msg, "handle_id")
            if !hid.Exists() {
                transferServerMsg(tid, msg, sess)
                break
            }

            handle, ok := j.Handle(sid.Uint(), hid.Uint())
            if !ok {
                log.Printf("janus: can't find handle with id %d", hid.Uint())
                break
            }

            transferServerMsg(tid, msg, handle)
        case sendMsg = <-j.sendChan:
            conn.WriteJSON(sendMsg)
        }
    }
}

func (j *Janus) Send(msg interface{}) {
    j.sendChan <- msg
}

func (j *Janus) ConnectStatus() bool {
    return j.connected
}

func (j *Janus) MsgChan(tid string) (msgChan chan []byte, exist bool) {
    return j.sessTable.MsgChan(tid)
}

func (j *Janus) NewSess(id uint64) *Session {
    return j.sessTable.newSess(id)
}

func (j *Janus) Session(id uint64) (sess *Session, exist bool) {
    sess, exist = j.sessTable.sessions[id]
    return sess, exist
}

func (j *Janus) Handle(sid uint64, hid uint64) (*Handle, bool) {
    sess, ok := j.sessTable.sessions[sid]
    if !ok {
        return nil, ok
    }

    handle, ok := sess.handles[hid]
    return handle, ok
}

func (j *Janus) NewTransaction() string {
    return j.sessTable.newTransaction()
}
