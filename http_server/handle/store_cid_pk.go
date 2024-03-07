package handle

import (
	"das-multi-device/tables"
	"encoding/hex"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/sign"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqStoreCidPk struct {
	Cid       string `json:"cid" binding:"required"`
	SignAddr  string `json:"sign_addr" binding:"required"`
	Msg       string `json:"msg" binding:"required"`
	Signature string `json:"signature" binding:"required"`
}

func (h *HttpHandle) StoreCidPk(ctx *gin.Context) {
	var (
		funcName = "StoreCidPk"
		clientIp = GetClientIp(ctx)
		req      *ReqStoreCidPk
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}

	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx)

	if err = h.doStoreCidPk(req, &apiResp); err != nil {
		log.Error("doStoreCidPk err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doStoreCidPk(req *ReqStoreCidPk, apiResp *http_api.ApiResp) (err error) {
	signType := common.DasAlgorithmIdWebauthn
	signAddressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     common.ChainTypeWebauthn,
		AddressNormal: req.SignAddr,
	})
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "sign address NormalToHex err")
		return err
	}
	h.dasCore.AddPkIndexForSignMsg(&req.Signature, 255)
	signMsg := req.Msg
	signature := req.Signature
	address := signAddressHex.AddressHex
	cid1 := common.CalculateCid1(req.Cid)
	if address[:20] != hex.EncodeToString(cid1) {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "cid err")
		return nil
	}
	verifyRes, _, err := http_api.VerifySignature(signType, signMsg, signature, address)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeSignError, "VerifySignature err")
		return fmt.Errorf("VerifySignature err: %s", err.Error())
	}
	if !verifyRes {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "Validation failed")
		return nil
	}
	var cidPk tables.TableCidPk
	cidPk.OriginPk = sign.GetPkFromSignature(common.Hex2Bytes(signature))
	cidPk.Pk = address[20:]
	cidPk.Cid = common.Bytes2Hex(cid1)
	if err := h.dbDao.InsertCidPk(cidPk); err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "InsertCidPk err")
		return fmt.Errorf("InsertCidPk err: %s", err.Error())
	}
	apiResp.ApiRespOK(true)
	return
}
