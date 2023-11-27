package handle

import (
	"context"
	"das-multi-device/cache"
	"das-multi-device/config"
	"das-multi-device/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/dascache"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/http_api/logger"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

var (
	log = logger.NewLogger("handle", logger.LevelDebug)
)

type HttpHandle struct {
	ctx                    context.Context
	dbDao                  *dao.DbDao
	rc                     *cache.RedisCache
	dasCore                *core.DasCore
	dasCache               *dascache.DasCache
	txBuilderBase          *txbuilder.DasTxBuilderBase
	mapReservedAccounts    map[string]struct{}
	mapUnAvailableAccounts map[string]struct{}
	serverScript           *types.Script
}

type HttpHandleParams struct {
	DbDao                  *dao.DbDao
	Rc                     *cache.RedisCache
	Ctx                    context.Context
	DasCore                *core.DasCore
	DasCache               *dascache.DasCache
	TxBuilderBase          *txbuilder.DasTxBuilderBase
	MapReservedAccounts    map[string]struct{}
	MapUnAvailableAccounts map[string]struct{}
	ServerScript           *types.Script
}

func Initialize(p HttpHandleParams) *HttpHandle {
	hh := HttpHandle{
		dbDao:                  p.DbDao,
		rc:                     p.Rc,
		ctx:                    p.Ctx,
		dasCore:                p.DasCore,
		dasCache:               p.DasCache,
		txBuilderBase:          p.TxBuilderBase,
		serverScript:           p.ServerScript,
		mapReservedAccounts:    p.MapReservedAccounts,
		mapUnAvailableAccounts: p.MapUnAvailableAccounts,
	}
	return &hh
}

func GetClientIp(ctx *gin.Context) string {
	clientIP := fmt.Sprintf("%v", ctx.Request.Header.Get("X-Real-IP"))
	return fmt.Sprintf("(%s)(%s)", clientIP, ctx.Request.RemoteAddr)
}

func (h *HttpHandle) checkSystemUpgrade(apiResp *http_api.ApiResp) error {
	if config.Cfg.Server.IsUpdate {
		apiResp.ApiRespErr(http_api.ApiCodeSystemUpgrade, "The service is under maintenance, please try again later.")
		return fmt.Errorf("backend system upgrade")
	}
	ConfigCellDataBuilder, err := h.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsMain)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "ConfigCellDataBuilderByTypeArgs err")
		return fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
	}
	status, _ := ConfigCellDataBuilder.Status()
	if status != 1 {
		apiResp.ApiRespErr(http_api.ApiCodeSystemUpgrade, "The service is under maintenance, please try again later.")
		return fmt.Errorf("contract system upgrade")
	}
	ok, err := h.dasCore.CheckContractStatusOK(common.DasContractNameAccountCellType)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "CheckContractStatusOK err")
		return fmt.Errorf("CheckContractStatusOK err: %s", err.Error())
	} else if !ok {
		apiResp.ApiRespErr(http_api.ApiCodeSystemUpgrade, "The service is under maintenance, please try again later.")
		return fmt.Errorf("contract system upgrade")
	}
	return nil
}
