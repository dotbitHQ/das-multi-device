package handle

import (
	"das-multi-device/http_server/api_code"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"sync"
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

func (h *HttpHandle) WebRTCWebSocket(ctx *gin.Context) {
	apiResp := api_code.ApiResp{}
	conn, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Error("websocket Upgrader err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "websocket Upgrader error")
		return
	}
	defer conn.Close()

	for {
		msg := &MsgData{}
		if err := conn.ReadJSON(msg); err != nil {
			log.Error("coon ReadJSON err:", err.Error())
			break
		}

		switch msg.Type {
		case "join":
			peersMapLock.Lock()
			PeersMap[msg.From] = conn
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
