// Copyright 2018 Cibifang. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
    "janus"
    "log"
    "net/http"
    "time"
    "net/url"

    "github.com/gorilla/websocket"
)

/* janus server addr */
var janusAddr = "www.janus_test.com:8188"
var janusUrl = url.URL{Scheme: "ws", Host: janusAddr, Path: "/janus"}

var localAddr = "localhost:10086"

var upgrader = websocket.Upgrader{}

func route(w http.ResponseWriter, r *http.Request) {
    log.Printf("receive request: %+v", r)
    // solve original check error
    r.Header.Del("Origin")
    c, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("upgrade:", err)
        return
    }
    defer c.Close()

    j := janus.NewJanus(janusUrl.String())
    go j.WaitMsg()

    wait := 0
    for j.ConnectStatus() == false {
        if wait >= 5 {
            log.Printf("wait %d s to connect, quit", wait)
            return
        }
        timer := time.NewTimer(time.Second * 1)
        <- timer.C
        wait += 1
        log.Printf("wait %d s to connect", wait)
    }

    var msg janus.ClientMsg
    var req janus.ServerMsg
    var tid string
    var reqChan chan janus.ServerMsg
    var ok bool

    for {
        if err := c.ReadJSON(&msg); err != nil {
            log.Println("c.ReadJSON error: ", err)
            break
        }
        log.Printf("recv: %+v", msg)
        switch msg.Janus {
        case "create":
            tid = j.NewTransaction()
            msg.Transaction = tid
            log.Println("create session")
            j.Send(msg)
            if reqChan, ok = j.MsgChan(tid); ok {
                req = <- reqChan
                log.Printf("receive from channel: %+v", req)
                j.NewSess(req.Data.Id)
                if err = c.WriteJSON(req); err != nil {
                    log.Println("WriteJSON error: ", err)
                }
            } else {
                log.Printf("create: can't find channel for tid %s", tid)
            }
        case "attach":
            sess, _ := j.Session(msg.SessionId)
            tid = sess.NewTransaction()
            msg.Transaction = tid
            log.Println("attach handle")
            j.Send(msg)
            if reqChan, ok = sess.MsgChan(tid); ok {
                req = <- reqChan
                log.Printf("receive from channel: %+v", req)
                sess.Attach(req.Data.Id)
                c.WriteJSON(req)
            } else {
                log.Printf("attach: can't find channel for tid %s", tid)
            }
        }
    }
}

func main() {
    http.HandleFunc("/ws", route)
    log.Fatal(http.ListenAndServe(localAddr, nil))
}
