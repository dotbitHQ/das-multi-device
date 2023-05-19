package http_server

import (
	"context"
	"das-multi-device/cache"
	"das-multi-device/dao"
	"das-multi-device/http_server/handle"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/dascache"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/mylog"
	"net/http"
)

var (
	log = mylog.NewLogger("http_server", mylog.LevelDebug)
)

type HttpServer struct {
	ctx             context.Context
	address         string
	engine          *gin.Engine
	srv             *http.Server
	internalAddress string
	internalEngine  *gin.Engine
	internalSrv     *http.Server
	h               *handle.HttpHandle
	rc              *cache.RedisCache
}

type HttpServerParams struct {
	Ctx             context.Context
	Address         string
	InternalAddress string
	DbDao           *dao.DbDao
	Rc              *cache.RedisCache
	DasCore         *core.DasCore
	DasCache        *dascache.DasCache
	TxBuilderBase   *txbuilder.DasTxBuilderBase
	ServerScript    *types.Script
}

func Initialize(p HttpServerParams) (*HttpServer, error) {
	hs := HttpServer{
		ctx:             p.Ctx,
		address:         p.Address,
		internalAddress: p.InternalAddress,
		engine:          gin.New(),
		internalEngine:  gin.New(),
		h: handle.Initialize(handle.HttpHandleParams{
			DbDao:         p.DbDao,
			Rc:            p.Rc,
			DasCore:       p.DasCore,
			Ctx:           p.Ctx,
			DasCache:      p.DasCache,
			TxBuilderBase: p.TxBuilderBase,
			ServerScript:  p.ServerScript,
		}),
		rc: p.Rc,
	}
	return &hs, nil
}

func (h *HttpServer) Run() {
	h.initRouter()
	h.srv = &http.Server{
		Addr:    h.address,
		Handler: h.engine,
	}
	h.internalSrv = &http.Server{
		Addr:    h.internalAddress,
		Handler: h.internalEngine,
	}
	go func() {
		if err := h.srv.ListenAndServe(); err != nil {
			log.Error("http_server run err:", err)
		}
	}()

	go func() {
		if err := h.internalSrv.ListenAndServe(); err != nil {
			log.Error("http_server internal run err:", err)
		}
	}()
}

func (h *HttpServer) Shutdown() {
	if h.srv != nil {
		log.Warn("http server Shutdown ... ")
		if err := h.srv.Shutdown(h.ctx); err != nil {
			log.Error("http server Shutdown err:", err.Error())
		}
	}
	if h.internalSrv != nil {
		log.Warn("http server internal Shutdown ... ")
		if err := h.internalSrv.Shutdown(h.ctx); err != nil {
			log.Error("http server internal Shutdown err:", err.Error())
		}
	}
}
