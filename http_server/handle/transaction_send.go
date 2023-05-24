package handle

import (
	"bytes"
	"das-multi-device/http_server/api_code"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"strings"
)

type ReqTransactionSend struct {
	SignInfo
}

type RespTransactionSend struct {
	Hash string `json:"hash"`
}

func (h *HttpHandle) RpcTransactionSend(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqTransactionSend
	err := json.Unmarshal(p, &req)
	if err != nil {
		log.Error("json.Unmarshal err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return
	} else if len(req) == 0 {
		log.Error("len(req) is 0")
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return
	}

	if err = h.doTransactionSend(&req[0], apiResp); err != nil {
		log.Error("doVersion err:", err.Error())
	}
}

func (h *HttpHandle) TransactionSend(ctx *gin.Context) {
	var (
		funcName = "TransactionSend"
		clientIp = GetClientIp(ctx)
		req      ReqTransactionSend
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

	if err = h.doTransactionSend(&req, &apiResp); err != nil {
		log.Error("doTransactionSend err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doTransactionSend(req *ReqTransactionSend, apiResp *api_code.ApiResp) error {
	var resp RespTransactionSend

	var sic SignInfoCache
	// get tx by cache
	if txStr, err := h.rc.GetSignTxCache(req.SignKey); err != nil {
		if err == redis.Nil {
			apiResp.ApiRespErr(api_code.ApiCodeTxExpired, "tx expired err")
		} else {
			apiResp.ApiRespErr(api_code.ApiCodeCacheError, "cache err")
		}
		return fmt.Errorf("GetSignTxCache err: %s", err.Error())
	} else if err = json.Unmarshal([]byte(txStr), &sic); err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "json.Unmarshal err")
		return fmt.Errorf("json.Unmarshal err: %s", err.Error())
	}

	hasWebAuthn := false
	for _, v := range req.SignList {
		if v.SignType == common.DasAlgorithmIdWebauthn {
			hasWebAuthn = true
			break
		}
	}
	if hasWebAuthn {
		webAuthnLockScript, _, err := h.dasCore.Daf().HexToScript(core.DasAddressHex{
			DasAlgorithmId: common.DasAlgorithmIdWebauthn,
			ChainType:      common.ChainTypeWebauthn,
			AddressHex:     sic.Address,
		})
		keyListConfigCellContract, err := core.GetDasContractInfo(common.DasKeyListConfigCellType)
		if err != nil {
			return fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		}
		searchKey := &indexer.SearchKey{
			Script:     webAuthnLockScript,
			ScriptType: indexer.ScriptTypeLock,
			Filter: &indexer.CellsFilter{
				Script: keyListConfigCellContract.ToScript(webAuthnLockScript.Args),
			},
		}
		res, err := h.dasCore.Client().GetCells(h.ctx, searchKey, indexer.SearchOrderDesc, 1, "")
		if err != nil {
			return fmt.Errorf("GetCells err: %s", err.Error())
		}
		if len(res.Objects) == 0 {
			return fmt.Errorf("can't find GetCells type: %s", common.DasKeyListConfigCellType)
		}
		keyListConfigTx, err := h.dasCore.Client().GetTransaction(h.ctx, res.Objects[0].OutPoint.TxHash)
		if err != nil {
			return err
		}
		webAuthnKeyListConfigBuilder, err := witness.WebAuthnKeyListDataBuilderFromTx(keyListConfigTx.Transaction, common.DataTypeNew)
		if err != nil {
			return err
		}
		dataBuilder := webAuthnKeyListConfigBuilder.WebAuthnKeyListData.AsBuilder()
		keyList := dataBuilder.Build()

		idx := -1
		for i := 0; i < int(keyList.Len()); i++ {
			if bytes.Equal(keyList.Get(uint(i)).Pubkey().RawData(), common.Hex2Bytes(sic.Address)[10:]) {
				idx = i
				break
			}
		}
		if idx == -1 {
			return errors.New("the current signing device is not in the authorized list")
		}

		for idx, v := range req.SignList {
			if v.SignType != common.DasAlgorithmIdWebauthn {
				continue
			}
			req.SignList[idx].SignMsg += fmt.Sprintf("%02d", idx)
		}
	}

	// sign
	txBuilder := txbuilder.NewDasTxBuilderFromBase(h.txBuilderBase, sic.BuilderTx)
	if err := txBuilder.AddSignatureForTx(req.SignList); err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "add signature fail")
		return fmt.Errorf("AddSignatureForTx err: %s", err.Error())
	}

	// send tx
	if hash, err := txBuilder.SendTransaction(); err != nil {
		if strings.Contains(err.Error(), "PoolRejectedDuplicatedTransaction") ||
			strings.Contains(err.Error(), "Dead(OutPoint(") ||
			strings.Contains(err.Error(), "Unknown(OutPoint(") ||
			(strings.Contains(err.Error(), "getInputCell") && strings.Contains(err.Error(), "not live")) {
			apiResp.ApiRespErr(api_code.ApiCodeRejectedOutPoint, err.Error())
			return fmt.Errorf("SendTransaction err: %s", err.Error())
		}
		if strings.Contains(err.Error(), "-102 in the page") {
			apiResp.ApiRespErr(api_code.ApiCodeOperationFrequent, "account frequency limit")
			return fmt.Errorf("SendTransaction err: %s", err.Error())
		}
		apiResp.ApiRespErr(api_code.ApiCodeError500, "send tx err:"+err.Error())
		return fmt.Errorf("SendTransaction err: %s", err.Error())
	} else {
		resp.Hash = hash.Hex()
		if sic.Address != "" {

			// operate limit
			_ = h.rc.SetApiLimit(sic.ChainType, sic.Address, sic.Action)
			//_ = h.rc.SetAccountLimit(sic.Account, time.Minute*2)
			//
			//// cache tx inputs
			//h.dasCache.AddCellInputByAction("", sic.BuilderTx.Transaction.Inputs)
			//// pending tx
			//pending := tables.TableRegisterPendingInfo{
			//	Account:        sic.Account,
			//	Action:         sic.Action,
			//	ChainType:      sic.ChainType,
			//	Address:        sic.Address,
			//	Capacity:       sic.Capacity,
			//	Outpoint:       common.OutPoint2String(hash.Hex(), 0),
			//	BlockTimestamp: uint64(time.Now().UnixNano() / 1e6),
			//}
			//if err = h.dbDao.CreatePending(&pending); err != nil {
			//	log.Error("CreatePending err: ", err.Error(), toolib.JsonString(pending))
			//}
		}
	}

	apiResp.ApiRespOK(resp)
	return nil
}
