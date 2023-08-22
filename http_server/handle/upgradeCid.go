package handle

import (
	"das-multi-device/config"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqCidInfo struct {
	Cid string `json:"cid" binding:"required"`
}

type RepCidInfo struct {
	IsOldCid   bool `json:"is_old_cid"`
	HasAccount bool `json:"has_account"`
}

func (h *HttpHandle) CidInfo(ctx *gin.Context) {
	var (
		funcName = "CidInfo"
		clientIp = GetClientIp(ctx)
		req      ReqCidInfo
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doCidInfo(&req, &apiResp); err != nil {
		log.Error("doCidInfo err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doCidInfo(req *ReqCidInfo, apiResp *http_api.ApiResp) (err error) {
	var resp RepCidInfo
	cid := req.Cid
	oldCid := config.Cfg.OldCid
	cid1, ok := oldCid[cid]
	resp.IsOldCid = ok
	if ok {
		num, err := h.dbDao.GetAccountInfoByCid(cid1)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeDbError, "GetAccountInfosByCid err:"+err.Error())
			return fmt.Errorf("SearchCidPk err: %s", err.Error())
		}
		if num > 0 {
			resp.HasAccount = true
		}
	}
	apiResp.ApiRespOK(resp)
	return nil
}

type ReqQueryCid struct {
	Account string `json:"account" binding:"required"`
}
type RepQueryCid struct {
	Cid string `json:"cid"`
}

func (h *HttpHandle) QueryCid(ctx *gin.Context) {
	var (
		funcName = "CidInfo"
		clientIp = GetClientIp(ctx)
		req      ReqQueryCid
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doQueryCid(&req, &apiResp); err != nil {
		log.Error("doCidInfo err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}
func (h *HttpHandle) doQueryCid(req *ReqQueryCid, apiResp *http_api.ApiResp) (err error) {
	var resp RepQueryCid
	account := req.Account
	acc, err := h.dbDao.GetAccountInfoByAccount(account)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "GetAccountInfoByAccount err")
		return fmt.Errorf("GetAccountInfoByAccount err: %s", err.Error())
	}
	if acc.Id == 0 || acc.OwnerChainType != common.ChainTypeWebauthn {
		apiResp.ApiRespErr(http_api.ApiCodeAccountNotExist, "account not found")
		return fmt.Errorf("GetAccountInfoByAccoun: account not found")
	}
	oldCid := config.Cfg.OldCid
	for k, v := range oldCid {
		if v == acc.Owner[:20] {
			resp.Cid = k
		}
	}
	apiResp.ApiRespOK(resp)
	return nil
}
