package handle

import (
	"crypto/elliptic"
	"crypto/sha256"
	"das-multi-device/http_server/api_code"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
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
		funcName = "ReportBusinessProcess"
		clientIp = GetClientIp(ctx)
		req      *ReqEcrecover
		apiResp  api_code.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}

	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doEcrecover(req, &apiResp); err != nil {
		log.Error("doEcrecover err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doEcrecover(req *ReqEcrecover, apiResp *api_code.ApiResp) (err error) {
	var resp RespEcrecover
	signData := req.SignData
	if len(signData) < 2 {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "Webauthn sign data must exceed 2")
		return
	}

	curve := elliptic.P256()
	var recoverData [2]common.RecoverData
	for i := 0; i < 2; i++ {
		authenticatorData, err := hex.DecodeString(signData[i].AuthenticatorData)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "AuthenticatorData  error")
			return fmt.Errorf("AuthenticatorData is error : %s", signData[i].AuthenticatorData)
		}
		clientDataJson, err := hex.DecodeString(signData[i].ClientDataJson)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "ClientDataJson  error")
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
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "Signature  error")
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
	fmt.Println("x: ", hex.EncodeToString(realPubkey.X.Bytes()), " ---- ", "y:", hex.EncodeToString(realPubkey.Y.Bytes()))

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
