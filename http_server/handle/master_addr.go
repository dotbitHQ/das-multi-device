package handle

import (
	"das-multi-device/http_server/api_code"
	"encoding/hex"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqGetMasters struct {
	Cid string `json:"cid" binding:"required"`
}
type RespGetMasters struct {
	CkbAddress []string `json:ckb_address`
}

func (h *HttpHandle) GetMasters(ctx *gin.Context) {
	var (
		funcName = "GetMasters"
		clientIp = GetClientIp(ctx)
		req      *ReqGetMasters
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

	if err = h.doGetMasters(req, &apiResp); err != nil {
		log.Error("doGetMasters err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doGetMasters(req *ReqGetMasters, apiResp *api_code.ApiResp) (err error) {
	var resp RespGetMasters
	cid := req.Cid
	cid1 := common.CaculateCid1(cid)
	authorizes, err := h.dbDao.GetMasters(hex.EncodeToString(cid1))
	ckbAddress := make([]string, 0)
	for _, v := range authorizes {
		masterCidBytes, err := hex.DecodeString(v.MasterCid)
		if err != nil {
			return err
		}
		masterPkBytes, err := hex.DecodeString(v.MasterPk)
		if err != nil {
			return err
		}
		addressNormal, err := h.dasCore.Daf().HexToNormal(core.DasAddressHex{
			DasAlgorithmId:    common.DasAlgorithmIdWebauthn,
			DasSubAlgorithmId: common.DasWebauthnSubAlgorithmIdES256,
			AddressHex:        common.CaculateWebauthnPayload(masterCidBytes, masterPkBytes),
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
