// Copyright 2018 Cibifang. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package janus

import (
    "net/http"
    "time"
    "log"

    "github.com/gorilla/websocket"
)

type Janus struct {
    *sessTable
    remoteServer    string
    connected       bool
    sendChan        chan interface{}
}

// type Path struct {
//     sessionId   string
//     handleId    string
// }

var janusDial = websocket.Dialer{
    Proxy:            http.ProxyFromEnvironment,
    HandshakeTimeout: 45 * time.Second,
    Subprotocols:     []string{"janus-protocol"},
}

func NewJanus(rs string) *Janus {
    return &Janus{remoteServer: rs,
                  sessTable: newSessTable(),
                  connected: false,
                  sendChan: make(chan interface{})}
}

// func Path(sid string, hid string) {
//     return Path{sessionId: sid, handleId: hid}
// }

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
    read := make(chan ServerMsg)

    go func() {
        var msg ServerMsg

        defer func() {
            closeChan <- struct{}{}
        }()
        for {
            if err := conn.ReadJSON(&msg); err != nil {
                log.Println("c.ReadJSON error: ", err)
                break
            }
            read <- msg
        }
    }()

    var msg ServerMsg
    var sendMsg interface{}

loop:
    for {
        select {
        case <- closeChan:
            break loop
        case msg = <- read:
            log.Printf("janus: receive message %+v", msg)
            tid := msg.Transaction
            sid := msg.SessionId
            if sid == 0 {
                // if msgChan, ok = j.MsgChan(tid); ok {
                //     msgChan <- msg
                // } else {
                //     log.Println("main: can't find channel with tid %s", tid)
                // }
                transferServerMsg(tid, msg, j)
                break
            }

            sess, ok := j.Session(sid)
            if  !ok {
                log.Printf("janus: can't find session with id %d", sid)
                break
            }

            hid := msg.HandleId
            if hid == 0 {
                transferServerMsg(tid, msg, sess)
                break
            }

            handle, ok := j.Handle(sid, hid)
            if !ok {
                log.Printf("janus: can't find handle with id %d", hid)
                break
            }

            // if msgChan, ok = sess.MsgChan()
            transferServerMsg(tid, msg, handle)
        case sendMsg = <- j.sendChan:
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

func (j *Janus) MsgChan(tid string) (msgChan (chan ServerMsg), exist bool) {
    return j.sessTable.MsgChan(tid)
}

// func (j *Janus) MainMsg(tid string) ((chan ServerMsg), ok) {
//     return j.sessTable.msgs[tid]
// }

// func (j *Janus) SessMsg(sid string, tid string) ((chan ServerMsg), ok) {
//     if sess, ok := j.sessTable.sessions[sid]; !ok {
//         return nil, ok
//     }

//     if _, ok = sess.msgs[tid]; !ok {
//         return nil, ok
//     }

//     return sess.msgs[tid]
// }

func (j *Janus) NewSess(id int) *Session {
    return j.sessTable.newSess(id)
}

func (j *Janus) Session(id int) (sess *Session, exist bool) {
    sess, exist = j.sessTable.sessions[id]
    return sess, exist
}

func (j *Janus) Handle(sid int, hid int) (*Handle, bool) {
    sess, ok := j.sessTable.sessions[sid];
    if !ok {
        return nil, ok
    }

    handle, ok := sess.handles[hid]
    return handle, ok
}

func (j *Janus) NewTransaction() string {
    return j.sessTable.newTransaction()
}
