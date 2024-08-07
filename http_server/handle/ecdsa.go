package handle

import (
	"context"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/asn1"
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

type WebauthnSignData struct {
	AuthenticatorData string `json:"authenticatorData"`
	ClientDataJson    string `json:"clientDataJson"`
	Signature         string `json:"signature"`
}

type ReqEcrecover struct {
	Cid      string              `json:"cid" binding:"required"`
	SignData []*WebauthnSignData `json:"sign_data" binding:"required"`
}
type RespEcrecover struct {
	CkbAddress string `json:"ckb_address"`
}

func (h *HttpHandle) Ecrecover(ctx *gin.Context) {
	var (
		funcName = "Ecrecover"
		clientIp = GetClientIp(ctx)
		req      *ReqEcrecover
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

	if err = h.doEcrecover(ctx.Request.Context(), req, &apiResp); err != nil {
		log.Error("doEcrecover err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doEcrecover(ctx context.Context, req *ReqEcrecover, apiResp *http_api.ApiResp) (err error) {
	var resp RespEcrecover
	signData := req.SignData
	if len(signData) < 2 {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "Webauthn sign data must exceed 2")
		return
	}

	curve := elliptic.P256()
	var recoverData [2]common.RecoverData
	for i := 0; i < 2; i++ {
		authenticatorData, err := hex.DecodeString(signData[i].AuthenticatorData)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "AuthenticatorData  error")
			return fmt.Errorf("AuthenticatorData is error : %s", signData[i].AuthenticatorData)
		}
		clientDataJson, err := hex.DecodeString(signData[i].ClientDataJson)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "ClientDataJson  error")
			return fmt.Errorf("ClientDataJson is error : %s", signData[i].ClientDataJson)
		}
		clientDataJsonHash := sha256.Sum256(clientDataJson)
		msg := append(authenticatorData, clientDataJsonHash[:]...)
		hash := sha256.Sum256(msg)
		//signature
		type ECDSASignature struct {
			R, S *big.Int
		}

		signature, err := hex.DecodeString(signData[i].Signature)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "Signature  error")
			return fmt.Errorf("signature is error : %s", signData[i].Signature)
		}

		e := &ECDSASignature{}

		_, err = asn1.Unmarshal(signature, e)
		if err != nil {
			return fmt.Errorf("Error asn1 unmarshal signature %s:", err)
		}
		recoverData[i].SignDigest = hash[:]
		recoverData[i].R = e.R
		recoverData[i].S = e.S
	}
	realPubkey, err := common.EcdsaRecover(curve, recoverData)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Info(ctx, "x: ", hex.EncodeToString(realPubkey.X.Bytes()), " ---- ", "y:", hex.EncodeToString(realPubkey.Y.Bytes()))

	normalAddress, err := h.dasCore.Daf().HexToNormal(core.DasAddressHex{
		DasAlgorithmId:    common.DasAlgorithmIdWebauthn,
		DasSubAlgorithmId: common.DasWebauthnSubAlgorithmIdES256,
		AddressHex:        common.GetWebauthnPayload(req.Cid, realPubkey),
	})
	if err != nil {
		return fmt.Errorf("HexToNormal err: %s", err.Error())
	}
	resp.CkbAddress = normalAddress.AddressNormal

	apiResp.ApiRespOK(resp)
	return nil
}

type ReqVerify struct {
	MasterAddr string `json:"master_addr" binding:"required"`
	BackupAddr string `json:"backup_addr" binding:"required"`
	Msg        string `json:"msg" binding:"required"`
	Signature  string `json:"signature" binding:"required"`
}
type RepVerify struct {
	IsValid bool `json:"is_valid"`
}

func (h *HttpHandle) VerifyWebauthnSign(ctx *gin.Context) {
	var (
		funcName = "VerifyWebauthnSign"
		clientIp = GetClientIp(ctx)
		req      *ReqVerify
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

	if err = h.doVerifyWebauthnSign(ctx.Request.Context(), req, &apiResp); err != nil {
		log.Error("doVerifyWebauthnSign err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doVerifyWebauthnSign(ctx context.Context, req *ReqVerify, apiResp *http_api.ApiResp) (err error) {
	var resp RepVerify
	signType := common.DasAlgorithmIdWebauthn
	backupAddressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     common.ChainTypeWebauthn,
		AddressNormal: req.BackupAddr,
	})
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "sign address NormalToHex err")
		return err
	}
	idx := 255
	if req.MasterAddr != req.BackupAddr {
		masterAddressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
			ChainType:     common.ChainTypeWebauthn,
			AddressNormal: req.MasterAddr,
		})
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "sign address NormalToHex err")
			return err
		}
		log.Info(ctx, masterAddressHex.AddressHex, "--", backupAddressHex.AddressHex)
		idx, err = h.dasCore.GetIdxOfKeylist(masterAddressHex, backupAddressHex)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "GetIdxOfKeylist err")
			return fmt.Errorf("GetIdxOfKeylist err: %s", err.Error())
		}
		if idx == -1 {
			apiResp.ApiRespErr(http_api.ApiCodePermissionDenied, "permission denied")
			return fmt.Errorf("permission denied")
		}
	}
	h.dasCore.AddPkIndexForSignMsg(&req.Signature, idx)
	signMsg := req.Msg
	signature := req.Signature
	address := backupAddressHex.AddressHex
	verifyRes, _, err := http_api.VerifySignature(signType, signMsg, signature, address)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeSignError, "VerifySignature err")
		return fmt.Errorf("VerifySignature err: %s", err.Error())
	}
	resp.IsValid = verifyRes
	apiResp.ApiRespOK(resp)
	return
}
