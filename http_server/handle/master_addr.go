package handle

import (
	"das-multi-device/config"
	"das-multi-device/http_server/api_code"
	"encoding/hex"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/address"
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
		payload := common.CaculateWebauthnPayload(masterCidBytes, masterPkBytes)
		addressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
			ChainType:     common.ChainTypeWebauthn,
			AddressNormal: payload,
			Is712:         true,
		})

		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "HexToArgs err")
			return fmt.Errorf("NormalToHex err: %s", err.Error())
		}

		lockScript, _, err := h.dasCore.Daf().HexToScript(addressHex)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
			return fmt.Errorf("HexToScript err: %s", err.Error())
		}

		if config.Cfg.Server.Net == common.DasNetTypeMainNet {
			addr, err := address.ConvertScriptToAddress(address.Mainnet, lockScript)
			if err != nil {
				apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
				return fmt.Errorf("ConvertScriptToAddress err: %s", err.Error())
			}

			ckbAddress = append(ckbAddress, addr)
		} else {
			addr, err := address.ConvertScriptToAddress(address.Testnet, lockScript)
			if err != nil {
				apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
				return fmt.Errorf("ConvertScriptToAddress err: %s", err.Error())
			}
			ckbAddress = append(ckbAddress, addr)
		}
	}
	resp.CkbAddress = ckbAddress
	apiResp.ApiRespOK(resp)
	return nil
}
