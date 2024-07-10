package handle

import (
	"context"
	"das-multi-device/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqAddCidInfo struct {
	CkbAddr string `json:"ckb_addr" binding:"required"`
	Cid     string `json:"cid" binding:"required"`
	Notes   string `json:"notes" binding:"required"`
	Device  string `json:"device" binding:"required"`
}

func (h *HttpHandle) AddCidInfo(ctx *gin.Context) {
	var (
		funcName = "AddCidInfo"
		clientIp = GetClientIp(ctx)
		req      ReqAddCidInfo
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx.Request.Context())

	if err = h.doAddCidInfo(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doAddCidInfo err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doAddCidInfo(ctx context.Context, req *ReqAddCidInfo, apiResp *http_api.ApiResp) (err error) {
	masterAddressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     common.ChainTypeWebauthn,
		AddressNormal: req.CkbAddr,
	})
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "NormalToHex err")
		return err
	}
	cid1 := common.Bytes2Hex(masterAddressHex.AddressPayload[:10])
	//Check if cid is enabled keyListConfigCell
	log.Info(ctx, "cid1: ", cid1)
	if common.Bytes2Hex(common.CalculateCid1(req.Cid)) != cid1 {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "ckb_addr or cid is error")
		return
	}

	notes := req.Notes
	device := req.Device
	if len(notes) > 40 || len(device) > 40 {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "notes or device error")
		return
	}

	cidInfo := tables.TableCidInfo{
		OriginalCid: req.Cid,
		Cid:         cid1,
		Notes:       notes,
		Device:      device,
	}
	res, err := h.dbDao.GetCidInfo(cid1)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "search cid_info err")
		return fmt.Errorf("CreateCidInfo err: %s", err.Error())
	}
	if res.Id > 0 {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "Cid already exists")
		return
	}
	if err := h.dbDao.CreateCidInfo(cidInfo); err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "create cid_info err")
		return fmt.Errorf("CreateCidInfo err: %s", err.Error())
	}
	apiResp.ApiRespOK(true)
	return nil
}
