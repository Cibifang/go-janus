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
	conn         *websocket.Conn
	sendChan     chan interface{}
	close        bool
}

var janusDial = websocket.Dialer{
	Proxy:            http.ProxyFromEnvironment,
	HandshakeTimeout: 45 * time.Second,
	Subprotocols:     []string{"janus-protocol"},
}

func NewJanus(rs string) *Janus {
	janus := &Janus{
		remoteServer: rs,
		sessTable:    newSessTable(),
		sendChan:     make(chan interface{}),
		close:        true,
	}

	if err := janus.connect(); err != nil {
		log.Printf(
			"NewJanus: connect to server `%s` failed, err : `%+v`", rs, err)
		return nil
	}

	janus.close = false
	go janus.handleMsg()
	return janus
}

func (j *Janus) Close() {
	if j.close {
		return
	}
	j.close = true
	j.conn.Close()
}

func (j *Janus) connect() error {
	conn, _, err := janusDial.Dial(j.remoteServer, nil)
	j.conn = conn
	return err
}

func (j *Janus) handleMsg() {
	readErr := make(chan struct{})
	read := make(chan []byte)

	go func() {
		defer j.Close()
		defer close(readErr)

		for {
			_, msg, err := j.conn.ReadMessage()
			if err != nil {
				log.Println("handleMsg: c.ReadJSON error: ", err)
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
		case <-readErr:
			break loop
		case msg = <-read:
			log.Printf("handleMsg: receive message `%s`", msg)
			tid := gjson.GetBytes(msg, "transaction").String()
			sid := gjson.GetBytes(msg, "session_id")
			if !sid.Exists() {
				transferServerMsg(tid, msg, j)
				break
			}

			sess, ok := j.Session(sid.Uint())
			if !ok {
				log.Printf(
					"handleMsg: can't find session with id `%d`", sid.Uint())
				break
			}

			hid := gjson.GetBytes(msg, "handle_id")
			if !hid.Exists() {
				transferServerMsg(tid, msg, sess)
				break
			}

			handle, ok := j.Handle(sid.Uint(), hid.Uint())
			if !ok {
				log.Printf(
					"handleMsg: can't find handle with id `%d`", hid.Uint())
				break
			}

			transferServerMsg(tid, msg, handle)
		case sendMsg = <-j.sendChan:
			log.Printf("handleMsg: send message `%+v`", sendMsg)
			j.conn.WriteJSON(sendMsg)
		}
	}
}

func (j *Janus) Send(msg interface{}) {
	j.sendChan <- msg
}

func (j *Janus) MsgChan(tid string) (msgChan chan []byte, exist bool) {
	return j.sessTable.MsgChan(tid)
}

func (j *Janus) DefaultMsgChan() chan []byte {
	return j.sessTable.DefaultMsgChan()
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
