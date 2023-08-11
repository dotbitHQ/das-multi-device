package handle

import (
	"das-multi-device/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"strings"
	"time"
)

type ReqTransactionSend struct {
	SignInfo
}

type RespTransactionSend struct {
	Hash string `json:"hash"`
}
type ReqTransactionStatus struct {
	ChainType common.ChainType  `json:"chain_type"`
	Address   string            `json:"address"`
	Actions   []tables.TxAction `json:"actions"`
	TxHash    string            `json:"tx_hash"`
}

func (h *HttpHandle) RpcTransactionSend(p json.RawMessage, apiResp *http_api.ApiResp) {
	var req []ReqTransactionSend
	err := json.Unmarshal(p, &req)
	if err != nil {
		log.Error("json.Unmarshal err:", err.Error())
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		return
	} else if len(req) == 0 {
		log.Error("len(req) is 0")
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
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

	if err = h.doTransactionSend(&req, &apiResp); err != nil {
		log.Error("doTransactionSend err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doTransactionSend(req *ReqTransactionSend, apiResp *http_api.ApiResp) error {
	var resp RespTransactionSend

	var sic SignInfoCache
	// get tx by cache
	if txStr, err := h.rc.GetSignTxCache(req.SignKey); err != nil {
		if err == redis.Nil {
			apiResp.ApiRespErr(http_api.ApiCodeTxExpired, "tx expired err")
		} else {
			apiResp.ApiRespErr(http_api.ApiCodeCacheError, "cache err")
		}
		return fmt.Errorf("GetSignTxCache err: %s", err.Error())
	} else if err = json.Unmarshal([]byte(txStr), &sic); err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "json.Unmarshal err")
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
		keyListCfgOutPoint := common.String2OutPointStruct(sic.KeyListCfgCellOpt)
		signAddressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
			ChainType:     common.ChainTypeWebauthn,
			AddressNormal: req.SignAddress, //Signed address
		})
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "sign address NormalToHex err")
			return err
		}
		idx, err := h.dasCore.GetIdxOfKeylistByOutPoint(keyListCfgOutPoint, signAddressHex)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeError500, "GetIdxOfKeylistByOp err: "+err.Error())
			return err
		}
		if idx == -1 {
			apiResp.ApiRespErr(http_api.ApiCodePermissionDenied, "permission denied")
			return fmt.Errorf("permission denied")
		}
		log.Info("signAddr loginAddr: ", signAddressHex.AddressHex, sic.Address)
		for i, v := range req.SignList {
			if v.SignType != common.DasAlgorithmIdWebauthn {
				continue
			}
			h.dasCore.AddPkIndexForSignMsg(&req.SignList[i].SignMsg, idx)
		}
	} else {
		//todo warning 日志
	}

	// sign
	txBuilder := txbuilder.NewDasTxBuilderFromBase(h.txBuilderBase, sic.BuilderTx)
	if err := txBuilder.AddSignatureForTx(req.SignList); err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "add signature fail")
		return fmt.Errorf("AddSignatureForTx err: %s", err.Error())
	}

	// send tx
	if hash, err := txBuilder.SendTransaction(); err != nil {
		if strings.Contains(err.Error(), "PoolRejectedDuplicatedTransaction") ||
			strings.Contains(err.Error(), "Dead(OutPoint(") ||
			strings.Contains(err.Error(), "Unknown(OutPoint(") ||
			(strings.Contains(err.Error(), "getInputCell") && strings.Contains(err.Error(), "not live")) {
			apiResp.ApiRespErr(http_api.ApiCodeRejectedOutPoint, err.Error())
			return fmt.Errorf("SendTransaction err: %s", err.Error())
		}
		if strings.Contains(err.Error(), "-102 in the page") {
			apiResp.ApiRespErr(http_api.ApiCodeOperationFrequent, "account frequency limit")
			return fmt.Errorf("SendTransaction err: %s", err.Error())
		}
		apiResp.ApiRespErr(http_api.ApiCodeError500, "send tx err:"+err.Error())
		return fmt.Errorf("SendTransaction err: %s", err.Error())
	} else {
		resp.Hash = hash.Hex()
		if sic.Address != "" {

			// operate limit
			_ = h.rc.SetApiLimit(sic.ChainType, sic.Address, sic.Action)

			pending := tables.TableWebauthnPendingInfo{
				Action:         sic.Action,
				ChainType:      sic.ChainType,
				Address:        sic.Address,
				Capacity:       sic.Capacity,
				Outpoint:       common.OutPoint2String(hash.Hex(), 0),
				BlockTimestamp: uint64(time.Now().UnixNano() / 1e6),
			}
			if err = h.dbDao.CreatePending(&pending); err != nil {
				log.Error("CreatePending err: ", err.Error(), toolib.JsonString(pending))
			}
		}
	}

	apiResp.ApiRespOK(resp)
	return nil
}

func (h *HttpHandle) TransactionStatus(ctx *gin.Context) {
	var (
		funcName = "TransactionStatus"
		clientIp = GetClientIp(ctx)
		req      ReqTransactionStatus
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

	if err = h.doTransactionStatus(&req, &apiResp); err != nil {
		log.Error("doTransactionStatus err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

type RespTransactionStatus struct {
	BlockNumber uint64          `json:"block_number"`
	Hash        string          `json:"hash"`
	Action      tables.TxAction `json:"action"`
	Status      int             `json:"status"`
}

func (h *HttpHandle) doTransactionStatus(req *ReqTransactionStatus, apiResp *http_api.ApiResp) error {
	//addressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
	//	ChainType:     req.ChainType,
	//	AddressNormal: req.Address,
	//	Is712:         true,
	//})
	//if err != nil {
	//	apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address NormalToHex err")
	//	return fmt.Errorf("NormalToHex err: %s", err.Error())
	//}
	//req.ChainType, req.Address = addressHex.ChainType, addressHex.AddressHex

	var resp RespTransactionStatus
	actionList := make([]common.DasAction, 0)
	for _, v := range req.Actions {
		actionList = append(actionList, tables.FormatActionType(v))
	}
	var tx tables.TableWebauthnPendingInfo
	var err error
	if req.TxHash != "" {
		tx, err = h.dbDao.GetTxStatusByOutpoint(req.TxHash)
	} else {
		tx, err = h.dbDao.GetTxStatus(req.ChainType, req.Address, actionList)
	}

	if err != nil && err.Error() != "record not found" {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "search tx status err")
		return fmt.Errorf("GetTransactionStatus err: %s", err.Error())
	}
	if tx.Id == 0 {
		apiResp.ApiRespErr(http_api.ApiCodeTransactionNotExist, "not exits tx")
		return nil
	}
	resp.BlockNumber = tx.BlockNumber
	resp.Hash, _ = common.String2OutPoint(tx.Outpoint)
	resp.Action = tables.FormatTxAction(tx.Action)
	resp.Status = tx.Status

	apiResp.ApiRespOK(resp)
	return nil
}
