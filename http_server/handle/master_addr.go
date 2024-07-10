package handle

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"math/big"
	"net/http"
)

type ReqGetMasters struct {
	Cid string `json:"cid" binding:"required"`
}

type RespGetMasters struct {
	CkbAddress []string `json:"ckb_address"`
}

type RespGetOringinPk struct {
	OriginalPk string `json:"origin_pk"`
}

type ReqCaculateCkbAddr struct {
	Cid    string `json:"cid" binding:"required"`
	Pubkey struct {
		X string `json:"x" binding:"required"`
		Y string `json:"y" binding:"required"`
	} `json:"pubkey" binding:"required"`
}
type RespCaculateCkbAddr struct {
	CkbAddress string `json:"ckb_address"`
}

func (h *HttpHandle) GetMasters(ctx *gin.Context) {
	var (
		funcName = "GetMasters"
		clientIp = GetClientIp(ctx)
		req      *ReqGetMasters
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx.Request.Context())
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}

	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx.Request.Context())

	if err = h.doGetMasters(ctx.Request.Context(), req, &apiResp); err != nil {
		log.Error("doGetMasters err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doGetMasters(ctx context.Context, req *ReqGetMasters, apiResp *http_api.ApiResp) (err error) {
	var resp RespGetMasters
	cid := req.Cid
	cid1 := common.CalculateCid1(cid)
	authorizes, err := h.dbDao.GetMasters(common.Bytes2Hex(cid1))
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "getMaster err")
		return fmt.Errorf("GetMasters err :%s", err.Error())
	}
	ckbAddress := make([]string, 0)
	for _, v := range authorizes {
		masterCidBytes := common.Hex2Bytes(v.MasterCid)

		masterPkBytes := common.Hex2Bytes(v.MasterPk)

		addressNormal, err := h.dasCore.Daf().HexToNormal(core.DasAddressHex{
			DasAlgorithmId:    common.DasAlgorithmIdWebauthn,
			DasSubAlgorithmId: common.DasWebauthnSubAlgorithmIdES256,
			AddressHex:        common.CalculateWebauthnPayload(masterCidBytes, masterPkBytes),
		})
		if err != nil {
			return err
		}
		ckbAddress = append(ckbAddress, addressNormal.AddressNormal)
	}
	resp.CkbAddress = ckbAddress
	apiResp.ApiRespOK(resp)
	return nil
}

func (h *HttpHandle) GetOriginalPk(ctx *gin.Context) {
	var (
		funcName = "GetOriginalPk"
		clientIp = GetClientIp(ctx)
		req      *ReqGetMasters
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx.Request.Context())
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}

	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx.Request.Context())

	if err = h.doGetOriginalPk(ctx.Request.Context(), req, &apiResp); err != nil {
		log.Error("doGetoriginalPk err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doGetOriginalPk(ctx context.Context, req *ReqGetMasters, apiResp *http_api.ApiResp) (err error) {
	var resp RespGetOringinPk
	cid := req.Cid
	cid1 := common.CalculateCid1(cid)
	cidPk, err := h.dbDao.GetCidPk(common.Bytes2Hex(cid1))
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "GetCidPk err")
		return fmt.Errorf("GetMasters err :%s", err.Error())
	}
	resp.OriginalPk = cidPk.OriginPk
	apiResp.ApiRespOK(resp)
	return nil
}

func (h *HttpHandle) CalculateCkbaddr(ctx *gin.Context) {
	var (
		funcName = "CalculateCkbaddr"
		clientIp = GetClientIp(ctx)
		req      *ReqCaculateCkbAddr
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

	if err = h.doCalculateCkbAddr(req, &apiResp); err != nil {
		log.Error("doGetMasters err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doCalculateCkbAddr(req *ReqCaculateCkbAddr, apiResp *http_api.ApiResp) (err error) {
	var resp RespCaculateCkbAddr
	curve := elliptic.P256()
	pubkey := new(ecdsa.PublicKey)
	pubkey.Curve = curve
	xBytes, err := hex.DecodeString(req.Pubkey.X)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		return err
	}
	yBytes, err := hex.DecodeString(req.Pubkey.Y)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		return err
	}
	pubkey.X = new(big.Int).SetBytes(xBytes)
	pubkey.Y = new(big.Int).SetBytes(yBytes)
	normalAddress, err := h.dasCore.Daf().HexToNormal(core.DasAddressHex{
		DasAlgorithmId:    common.DasAlgorithmIdWebauthn,
		DasSubAlgorithmId: common.DasWebauthnSubAlgorithmIdES256,
		AddressHex:        common.GetWebauthnPayload(req.Cid, pubkey),
	})
	if err != nil {
		return fmt.Errorf("HexToNormal err: %s", err.Error())
	}
	resp.CkbAddress = normalAddress.AddressNormal

	apiResp.ApiRespOK(resp)
	return nil
}
