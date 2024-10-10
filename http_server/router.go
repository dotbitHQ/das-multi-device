package http_server

import (
	"das-multi-device/config"
	"das-multi-device/http_server/api_code"
	"das-multi-device/http_server/webrtc"
	"encoding/json"
	"github.com/dotbitHQ/das-lib/http_api"
	sentrygin "github.com/getsentry/sentry-go/gin"
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
	h.engine.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))
	h.engine.Use(http_api.ReqIdMiddleware())
	v1 := h.engine.Group("v1")
	{
		//shortExpireTime, longExpireTime, lockTime := time.Second*5, time.Second*15, time.Minute
		//shortDataTime, longDataTime := time.Minute*3, time.Minute*10
		//cacheHandleShort := toolib.MiddlewareCacheByRedis(h.rc.GetRedisClient(), false, shortDataTime, lockTime, shortExpireTime, respHandle)
		//cacheHandleLong := toolib.MiddlewareCacheByRedis(h.rc.GetRedisClient(), false, longDataTime, lockTime, longExpireTime, respHandle)

		v1.POST("/webauthn/ecdsa-ecrecover", api_code.DoMonitorLog("ecdsa-ecrecover"), h.h.Ecrecover)
		v1.POST("/webauthn/get-masters-addr", api_code.DoMonitorLog("get-masters-addr"), h.h.GetMasters)
		v1.POST("/webauthn/get-original-pk", api_code.DoMonitorLog("get-original-pk"), h.h.GetOriginalPk)
		v1.POST("/webauthn/authorize", api_code.DoMonitorLog("authorize"), h.h.Authorize)
		v1.POST("/webauthn/authorize-info", api_code.DoMonitorLog("authorize-info"), h.h.AuthorizeInfo)
		//v1.POST("/webauthn/calculate-ckbaddr", api_code.DoMonitorLog(api_code.MethodTransactionStatus), h.h.CaculateCkbaddr)
		v1.POST("/transaction/send", api_code.DoMonitorLog("transaction-send"), h.h.TransactionSend)
		v1.POST("/transaction/status", api_code.DoMonitorLog("transaction-status"), h.h.TransactionStatus)
		v1.POST("/webauthn/query-cid", api_code.DoMonitorLog("query-cid"), h.h.QueryCid)
		v1.POST("/webauthn/add-cid-info", api_code.DoMonitorLog("add-cid-info"), h.h.AddCidInfo)

		v1.POST("/webauthn/cid-info", api_code.DoMonitorLog("cid-info"), h.h.CidInfo)
		v1.POST("/webauthn/add-test-cid", api_code.DoMonitorLog("add-test-cid"), h.h.AddTestCid)
		v1.POST("/webauthn/cover-cid", api_code.DoMonitorLog("cover-cid"), h.h.CoverCid)
		v1.POST("/webauthn/verify", api_code.DoMonitorLog("cid-info"), h.h.VerifyWebauthnSign)
		v1.StaticFS("/webrtc/chatroom", http.FS(webrtc.WebRTC))
		v1.GET("/webrtc/socket", h.h.WebRTCWebSocket)
		v1.GET("/test/jenkins", func(c *gin.Context) {
			c.JSON(200, "main--v1.0.0")
		})
	}
}

func respHandle(c *gin.Context, res string, err error) {
	if err != nil {
		log.Error("respHandle err:", err.Error())
		c.AbortWithStatusJSON(http.StatusOK, http_api.ApiRespErr(http.StatusInternalServerError, err.Error()))
	} else if res != "" {
		var respMap map[string]interface{}
		_ = json.Unmarshal([]byte(res), &respMap)
		c.AbortWithStatusJSON(http.StatusOK, respMap)
	}
}
