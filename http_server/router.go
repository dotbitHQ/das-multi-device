package http_server

import (
	"das-multi-device/config"
	"das-multi-device/http_server/api_code"
	"das-multi-device/http_server/webrtc"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

func (h *HttpServer) initRouter() {

	log.Info("initRouter:", len(config.Cfg.Origins))
	if len(config.Cfg.Origins) > 0 {
		toolib.AllowOriginList = append(toolib.AllowOriginList, config.Cfg.Origins...)
	}
	h.internalEngine.Use(toolib.MiddlewareCors())
	h.engine.Use(toolib.MiddlewareCors())

	v1 := h.engine.Group("v1")
	{
		//shortExpireTime, longExpireTime, lockTime := time.Second*5, time.Second*15, time.Minute
		//shortDataTime, longDataTime := time.Minute*3, time.Minute*10
		//cacheHandleShort := toolib.MiddlewareCacheByRedis(h.rc.GetRedisClient(), false, shortDataTime, lockTime, shortExpireTime, respHandle)
		//cacheHandleLong := toolib.MiddlewareCacheByRedis(h.rc.GetRedisClient(), false, longDataTime, lockTime, longExpireTime, respHandle)

		v1.POST("/webauthn/ecdsa-ecrecover", api_code.DoMonitorLog(api_code.MethodEcdsaRecover), h.h.Ecrecover)
		v1.POST("/webauthn/get-masters-addr", api_code.DoMonitorLog(api_code.MethodGetMasterAddr), h.h.GetMasters)
		v1.POST("/webauthn/get-original-pk", api_code.DoMonitorLog(api_code.MethodGetOriginPk), h.h.GetOriginalPk)
		v1.POST("/webauthn/authorize", api_code.DoMonitorLog(api_code.MethodAuthorize), h.h.Authorize)
		v1.POST("/webauthn/authorize-info", api_code.DoMonitorLog(api_code.MethodAuthorize), h.h.AuthorizeInfo)
		//v1.POST("/webauthn/calculate-ckbaddr", api_code.DoMonitorLog(api_code.MethodTransactionStatus), h.h.CaculateCkbaddr)
		v1.POST("/transaction/send", api_code.DoMonitorLog(api_code.MethodTransactionSend), h.h.TransactionSend)
		v1.POST("/transaction/status", api_code.DoMonitorLog(api_code.MethodTransactionStatus), h.h.TransactionStatus)
		v1.StaticFS("/webrtc/chatroom", http.FS(webrtc.WebRTC))
		v1.GET("/webrtc/socket", h.h.WebRTCWebSocket)
	}
}

// 11
func respHandle(c *gin.Context, res string, err error) {
	if err != nil {
		log.Error("respHandle err:", err.Error())
		c.AbortWithStatusJSON(http.StatusOK, api_code.ApiRespErr(http.StatusInternalServerError, err.Error()))
	} else if res != "" {
		var respMap map[string]interface{}
		_ = json.Unmarshal([]byte(res), &respMap)
		c.AbortWithStatusJSON(http.StatusOK, respMap)
	}
}
