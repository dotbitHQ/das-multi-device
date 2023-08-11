package handle

import (
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type MsgData struct {
	From   string      `json:"from"`
	To     string      `json:"to"`
	Answer bool        `json:"answer"`
	Type   string      `json:"type"`
	Data   interface{} `json:"data"`
}

var peersMapLock = sync.RWMutex{}
var PeersMap = map[string]*websocket.Conn{}
var Conn2Cid = map[*websocket.Conn]string{}

func (h *HttpHandle) WebRTCWebSocket(ctx *gin.Context) {
	apiResp := http_api.ApiResp{}
	conn, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Error("websocket Upgrader err:", err.Error())
		apiResp.ApiRespErr(http_api.ApiCodeError500, "websocket Upgrader error")
		return
	}
	exist := make(chan struct{})

	defer func() {
		_ = conn.Close()
		peersMapLock.Lock()
		cid := Conn2Cid[conn]
		delete(Conn2Cid, conn)
		delete(PeersMap, cid)
		peersMapLock.Unlock()
		close(exist)
	}()

	conn.SetPingHandler(nil)
	conn.SetPongHandler(nil)
	conn.SetCloseHandler(nil)

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			select {
			case <-exist:
				return
			default:
				_ = conn.WriteMessage(websocket.PingMessage, []byte("ping"))
			}
		}
	}()

	for {
		msg := &MsgData{}
		if err := conn.ReadJSON(msg); err != nil {
			log.Error("coon ReadJSON err:", err.Error())
			break
		}

		switch msg.Type {
		case "join":
			peersMapLock.Lock()
			wg := sync.WaitGroup{}
			wg.Add(len(PeersMap))
			for k, v := range PeersMap {
				cid := k
				peerConn := v
				go func() {
					defer wg.Done()
					if err := peerConn.WriteJSON(MsgData{
						To:   cid,
						Type: "peers",
						Data: map[string]interface{}{
							"peers": msg.From,
						},
					}); err != nil {
						log.Errorf("send joiner to others err: %s", err)
					}
				}()
			}
			wg.Wait()

			PeersMap[msg.From] = conn
			Conn2Cid[conn] = msg.From
			peersMapLock.Unlock()

			cids := make([]map[string]interface{}, 0)

			peersMapLock.RLock()
			for k := range PeersMap {
				if k == msg.From {
					continue
				}
				cids = append(cids, map[string]interface{}{
					"cid": k,
				})
			}
			peersMapLock.RUnlock()

			if len(cids) == 0 {
				continue
			}
			if err := conn.WriteJSON(MsgData{
				From: "",
				To:   msg.From,
				Type: "peers",
				Data: map[string]interface{}{
					"peers": cids,
				},
			}); err != nil {
				log.Error("coon WriteJSON err:", err.Error())
				break
			}
		case "offer", "answer", "candidate":
			peersMapLock.RLock()
			c, ok := PeersMap[msg.To]
			if !ok {
				peersMapLock.RUnlock()
				log.Error("cid:%s not exist", msg.To)
				continue
			}
			peersMapLock.RUnlock()
			if err := c.WriteJSON(msg); err != nil {
				log.Error("coon WriteJSON err:", err.Error())
				break
			}
		}
	}
}
