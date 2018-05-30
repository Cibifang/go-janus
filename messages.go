// Copyright 2018 Cibifang. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package janus

import (
	"log"
	"math/rand"
	"time"
)

type ClientBody struct {
	Request    string `json:"request,omitempty"`
	Room       uint64 `json:"room,omitempty"`
	Ptype      string `json:"ptype,omitempty"`
	Feed       uint64 `json:"feed,omitempty"`
	PrivateId  uint64 `json:"private_id,omitempty"`
	Display    string `json:"display,omitempty"`
	Audio      bool   `json:"audio,omitempty"`
	Video      bool   `json:"video,omitempty"`
	Audiocodec string `json:"audiocodec,omitempty"`
	Videocodec string `json:"videocodec,omitempty"`
}

type Jsep struct {
	Type string `json:"type,omitempty"`
	Sdp  string `json:"sdp,omitempty"`
}

// WebSocket message send to janus.
type ClientMsg struct {
	Janus       string     `json:"janus,omitempty"`
	Plugin      string     `json:"plugin,omitempty"`
	Transaction string     `json:"transaction,omitempty"`
	SessionId   uint64     `json:"session_id,omitempty"`
	HandleId    uint64     `json:"handle_id,omitempty"`
	Body        ClientBody `json:"body,omitempty"`
}

type ServerData struct {
	Id int `json:"id,omitempty"`
}

type ServerMsg struct {
	Janus       string     `json:"janus,omitempty"`
	Transaction string     `json:"transaction,omitempty"`
	Data        ServerData `json:"data,omitempty"`
	SessionId   int        `json:"session_id,omitempty"`
	HandleId    int        `json:"handle_id,omitempty"`
	Plugindata  struct {
		Plugin string `json:"plugin,omitempty"`
		Data   struct {
			Videoroom   string `json:"videoroom,omitempty"`
			Room        int64  `json:"room,omitempty"`
			Permanent   bool   `json:"permanent,omitempty"`
			Description string `json:"description,omitempty"`
			Id          int64  `json:"id,omitempty"`
			PrivateId   int64  `json:"private_id,omitempty"`
			Publishers  []struct {
				Id         int64  `json:"id,omitempty"`
				Display    string `json:"display,omitempty"`
				AudioCodec string `json:"audio_codec,omitempty"`
				VideoCodec string `json:"video_codec,omitempty"`
			} `json:"publishers,omitempty"`
		} `json:"data,omitempty"`
	} `json:"plugindata,omitempty"`
	Jsep struct {
		Type string `json:"type,omitempty"`
		Sdp  string `json:"sdp,omitempty"`
	} `json:"jsep,omitempty"`
}

type ServerMsgChan interface {
	MsgChan(string) (chan []byte, bool)
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

func transferServerMsg(tid string, msg []byte, serverMsgChan ServerMsgChan) {
	if msgChan, ok := serverMsgChan.MsgChan(tid); ok {
		msgChan <- msg
	} else {
		log.Printf("janus: can't find channel with tid `%s`", tid)
	}
}
