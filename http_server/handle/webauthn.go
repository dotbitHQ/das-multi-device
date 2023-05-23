package handle

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"das-multi-device/cache"
	"das-multi-device/config"
	"das-multi-device/http_server/api_code"
	"das-multi-device/tables"
	"das-multi-device/tool"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"math/big"
	"net/http"
	"time"
)

type WebauthnSignData struct {
	AuthenticatorData string `json:"authenticatorData"`
	ClientDataJson    string `json:"clientDataJson"`
	Signature         string `json:"signature"`
}
type ReqEcrecover struct {
	Cid      string              `json:"cid"`
	SignData []*WebauthnSignData `json:"sign_data"`
}
type RespEcrecover struct {
	CkbAddress string `json:"ckb_address"`
}

type RespReportBusinessProcess struct {
	ProcessId string `json:"process_id"`
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
		log.Error("doReportBusinessProcess err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doEcrecover(req *ReqEcrecover, apiResp *api_code.ApiResp) (err error) {
	var resp RespEcrecover
	if req.Cid == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "Cid is empty")
		return
	}
	signData := req.SignData
	if len(signData) < 2 {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "Webauthn sign data must exceed 2")
		return
	}

	var pubKeys []*ecdsa.PublicKey
	for i := 0; i < 2; i++ {
		authenticatorData, err := hex.DecodeString(signData[i].AuthenticatorData)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "AuthenticatorData  error")
			return fmt.Errorf("AuthenticatorData is error : ", signData[i].AuthenticatorData)
		}
		clientDataJson, err := hex.DecodeString(signData[i].ClientDataJson)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "ClientDataJson  error")
			return fmt.Errorf("ClientDataJson is error : ", signData[i].ClientDataJson)
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
			return fmt.Errorf("Signature is error : ", signData[i].Signature)
		}

		e := &ECDSASignature{}

		_, err = asn1.Unmarshal(signature, e)
		if err != nil {
			fmt.Println("Error asn1 unmarshal signature ", err)
		}
		possiblePubkey, err := tool.GetPubKey(hash[:], e.R, e.S)
		pubKeys = append(pubKeys, possiblePubkey[:]...)
	}
	fmt.Println("all pubkeys: ", pubKeys)
	var realPubkey *ecdsa.PublicKey
	for i := 0; i < 2; i++ {
		if pubKeys[i].Equal(pubKeys[2]) || pubKeys[i].Equal(pubKeys[3]) {
			realPubkey = pubKeys[i]
		}
	}
	if realPubkey == nil {
		return fmt.Errorf("recover faild")
	}
	fmt.Println("realpubkeys: ", realPubkey)

	//计算ckb地址
	//计算webauthn payload

	webauthnPayload := common.GetWebauthnPayload(req.Cid, realPubkey)
	fmt.Println("webauthnPayload ", webauthnPayload)
	addressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     common.ChainTypeWebauthn,
		AddressNormal: webauthnPayload,
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
		resp.CkbAddress = addr
	} else {
		addr, err := address.ConvertScriptToAddress(address.Testnet, lockScript)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
			return fmt.Errorf("ConvertScriptToAddress err: %s", err.Error())
		}
		resp.CkbAddress = addr
	}
	apiResp.ApiRespOK(resp)
	return nil
}

type ReqGetMasters struct {
	Cid string `json:"cid"`
}
type RespGetMasters struct {
	CkbAddress []string `json:ckb_address`
}

type ReqAuthorize struct {
	MasterCkbAddress string `json:"master_ckb_address"`
	SlaveCkbAddress  string `json:"slave_ckb_address"`
}
type RespAuthorize struct {
	SignInfo
}

func (h *HttpHandle) GetMasters(ctx *gin.Context) {
	var (
		funcName = "ReportBusinessProcess"
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
		log.Error("doReportBusinessProcess err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doGetMasters(req *ReqGetMasters, apiResp *api_code.ApiResp) (err error) {
	var resp RespGetMasters
	cid := req.Cid
	if cid == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "Cid is empty")
		return
	}
	cid1 := common.CaculateCid1(cid)
	authorizes, err := h.dbDao.GetMasters(hex.EncodeToString(cid1))
	var ckbAddress []string
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

func (h *HttpHandle) Authorize(ctx *gin.Context) {
	var (
		funcName = "SubAccountInit"
		clientIp = GetClientIp(ctx)
		req      ReqAuthorize
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

	if err = h.doAuthorize(&req, &apiResp); err != nil {
		log.Error("doEditOwner err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doAuthorize(req *ReqAuthorize, apiResp *api_code.ApiResp) (err error) {
	var resp RespAuthorize
	var keyListConfigCellOutPoint string
	//根据ckb地址获取cid
	master_addr := req.MasterCkbAddress
	slave_addr := req.SlaveCkbAddress
	if master_addr == "" || slave_addr == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "master_address or slave_address is empty")
		return
	}
	masterPayload, err := h.dasCore.Daf().AddrToWebauthnPayload(master_addr)
	masterPayloadHex := common.Bytes2Hex(masterPayload)
	cid1 := common.Bytes2Hex(masterPayload[:10])
	fmt.Println(22222, cid1)
	//2a022782ef1db916da06733bec8b7c4a3bf286e8
	//检测该cid是已经开启里备份
	res, err := h.dbDao.GetCidPk(cid1)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search cidpk err")
		return fmt.Errorf("SearchCidPk err: %s", err.Error())
	}
	keyListConfigCellOutPoint = res.Outpoint
	if res.Id == 0 || res.EnableAuthorize == tables.EnableAuthorizeOff {
		//检测是否可以开通备份（是否拥有.bit资产或者有ckb余额）
		canCreate, err := h.checkCanBeCreated(masterPayloadHex)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "check if can be created err")
			return fmt.Errorf("checkCanBeCreated err : %s", err.Error())
		}
		if !canCreate {
			apiResp.ApiRespErr(api_code.ApiCodeHasNoAccessToCreate, "master_address can`t enable authorize")
			return fmt.Errorf("master_address hasn`t enable authorize")
		}

		//开启备份（创建keyListConfigCell）
		keyListConfigCellOutPoint, err = h.createKeyListCfgCell(masterPayloadHex)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeCreateConfigCellFail, "create keyListConfigCell err")
			return err
		}

	}

	//添加备份-交易
	slavePayload, err := h.dasCore.Daf().AddrToWebauthnPayload(slave_addr)
	reqBuildWebauthnTx := reqBuildWebauthnTx{
		Action:          common.DasActionUpdateKeyList,
		ChainType:       common.ChainTypeWebauthn,
		keyListConfigOp: keyListConfigCellOutPoint,
		MasterPayLoad:   masterPayload,
		SlavePayload:    slavePayload,
		Capacity:        0,
	}

	txParams, err := h.buildAddAuthorizeTx(&reqBuildWebauthnTx)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "build tx err: "+err.Error())
		return fmt.Errorf("buildEditManagerTx err: %s", err.Error())
	}
	if si, err := h.buildWebauthnTx(&reqBuildWebauthnTx, txParams); err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "buildWebauthnTx tx err: "+err.Error())
		return fmt.Errorf("buildWebauthnTx: %s", err.Error())
	} else {
		resp.SignInfo = *si
	}

	apiResp.ApiRespOK(resp)
	return nil

}

type reqBuildWebauthnTx struct {
	Action          common.DasAction
	ChainType       common.ChainType
	keyListConfigOp string
	MasterPayLoad   []byte
	SlavePayload    []byte
	Capacity        uint64 `json:"capacity"`
}

func (h *HttpHandle) buildAddAuthorizeTx(req *reqBuildWebauthnTx) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams
	contractDas, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	configMain, err := core.GetDasContractInfo(common.DasContractNameConfigCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	keyListCfgCell, err := core.GetDasContractInfo(common.DasKeyListConfigCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}

	// inputs account cell
	keyListCfgOutPoint := common.String2OutPointStruct(req.keyListConfigOp)
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: keyListCfgOutPoint,
	})

	actionWitness, err := witness.GenActionDataWitness(common.DasActionUpdateKeyList, nil)
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	ownerHex := core.DasAddressHex{
		DasAlgorithmId: common.DasAlgorithmIdWebauthn,
		AddressHex:     common.Bytes2Hex(req.MasterPayLoad),
		IsMulti:        false,
		ChainType:      common.ChainTypeWebauthn,
	}
	lockArgs, err := h.dasCore.Daf().HexToArgs(ownerHex, ownerHex)
	if err != nil {
		return nil, fmt.Errorf("HexToArgs err: %s", err.Error())
	}

	res, err := h.dasCore.Client().GetTransaction(h.ctx, keyListCfgOutPoint.TxHash)
	if err != nil {
		return nil, fmt.Errorf("GetTransaction err: %s", err.Error())
	}
	capacity := res.Transaction.Outputs[keyListCfgOutPoint.Index].Capacity

	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: capacity,
		Lock:     contractDas.ToScript(lockArgs),
		Type:     nil,
	})

	//
	res, err = h.dasCore.Client().GetTransaction(h.ctx, keyListCfgOutPoint.TxHash)
	if err != nil {
		return nil, fmt.Errorf("GetTransaction err: %s", err.Error())
	}
	builder, err := witness.WebAuthnKeyListDataBuilderFromTx(res.Transaction, common.DataTypeNew)
	if err != nil {
		return nil, fmt.Errorf("WebAuthnKeyListDataBuilderFromTx err: %s", err.Error())
	}
	var s witness.WebauthnKey
	s.MinAlgId = uint8(common.DasAlgorithmIdWebauthn)
	s.SubAlgId = uint8(common.DasWebauthnSubAlgorithmIdES256)
	s.Cid = string(req.SlavePayload[:10])
	s.PubKey = string(req.SlavePayload[10:])
	klWitness, klData, err := builder.GenWitness(&witness.WebauchnKeyListCellParam{
		Action:             common.DasActionUpdateKeyList,
		OldIndex:           0,
		NewIndex:           0,
		AddWebauthnKeyList: s,
	})
	txParams.Witnesses = append(txParams.Witnesses, klWitness)
	txParams.OutputsData = append(txParams.OutputsData, klData)

	//cell deps
	txParams.CellDeps = append(txParams.CellDeps,
		contractDas.ToCellDep(),
		configMain.ToCellDep(),
		keyListCfgCell.ToCellDep(),
	)
	return &txParams, nil
}

func (h *HttpHandle) buildWebauthnTx(req *reqBuildWebauthnTx, txParams *txbuilder.BuildTransactionParams) (*SignInfo, error) {
	txBuilder := txbuilder.NewDasTxBuilderFromBase(h.txBuilderBase, nil)
	if err := txBuilder.BuildTransaction(txParams); err != nil {
		return nil, fmt.Errorf("txBuilder.BuildTransaction err: %s", err.Error())
	}

	var skipGroups []int
	switch req.Action {
	case common.DasActionCreateKeyList:

	case common.DasActionUpdateKeyList:

	}
	signList, err := txBuilder.GenerateDigestListFromTx(skipGroups)
	if err != nil {
		return nil, fmt.Errorf("txBuilder.GenerateDigestListFromTx err: %s", err.Error())
	}

	log.Info("buildTx:", txBuilder.TxString())

	var sic SignInfoCache
	sic.Action = req.Action
	sic.ChainType = req.ChainType
	sic.Address = common.Bytes2Hex(req.MasterPayLoad)

	sic.Capacity = req.Capacity
	sic.BuilderTx = txBuilder.DasTxBuilderTransaction
	signKey := sic.SignKey()
	cacheStr := toolib.JsonString(&sic)
	if err = h.rc.SetSignTxCache(signKey, cacheStr); err != nil {
		return nil, fmt.Errorf("SetSignTxCache err: %s", err.Error())
	}

	var si SignInfo
	si.SignKey = signKey
	si.SignList = signList

	return &si, nil
}

func (h *HttpHandle) checkCanBeCreated(payload string) (canCreate bool, err error) {
	//是否拥有.bit账号
	num, err := h.dbDao.GetAccountInfos(payload)
	if err != nil {
		return false, fmt.Errorf("GetAccountInfos err: %s", err.Error())
	}
	if num > 0 {
		return true, nil
	}

	//是否有ckb资产
	dasLockScript, _, err := h.dasCore.Daf().HexToScript(core.DasAddressHex{
		DasAlgorithmId: common.DasAlgorithmIdWebauthn,
		AddressHex:     payload,
		IsMulti:        false,
		ChainType:      common.ChainTypeWebauthn,
	})
	if err != nil {
		return false, fmt.Errorf("FormatAddressToDasLockScript err: %s", err.Error())
	}
	_, dasLockAmount, err := h.dasCore.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache:          nil,
		LockScript:        dasLockScript,
		CapacityNeed:      0,
		CapacityForChange: 0,
		SearchOrder:       indexer.SearchOrderDesc,
	})
	if err != nil {
		return false, fmt.Errorf("GetBalanceCells err: %s", err.Error())
	}

	return dasLockAmount > 0, nil
}

// 创建 keyListConfigCell
func (h *HttpHandle) createKeyListCfgCell(payload string) (outPoint string, err error) {

	if res := h.rc.LockWithRedis(common.ChainTypeWebauthn, payload, cache.CreateKeyListConfigCell, time.Minute*4); res != nil {
		return "", fmt.Errorf("createKeyListCfgCell LockWithRedis err :%s", err.Error())
	}
	return
}

func (h *HttpHandle) buildCreateKeyListCfgTx(webauthnPayload string) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	//outputs
	keyListCfgContract, err := core.GetDasContractInfo(common.DasKeyListConfigCellType)
	if err != nil {
		return nil, fmt.Errorf("")
	}
	//outputsData
	var keyListCfgData []byte
	txParams.OutputsData = append(txParams.OutputsData, keyListCfgData)

	addressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     common.ChainTypeWebauthn,
		AddressNormal: webauthnPayload,
		Is712:         true,
	})
	if err != nil {

		return nil, fmt.Errorf("NormalToHex err: %s", err.Error())
	}

	lockScript, _, err := h.dasCore.Daf().HexToScript(addressHex)
	if err != nil {
		return nil, fmt.Errorf("HexToScript err: %s", err.Error())
	}
	//outputsCell
	keylistCfgOutput := &types.CellOutput{
		Lock: lockScript,
		Type: keyListCfgContract.ToScript(nil),
	}
	keylistCfgOutput.Capacity = keylistCfgOutput.OccupiedCapacity(keyListCfgData) * common.OneCkb
	txParams.Outputs = append(txParams.Outputs, keylistCfgOutput)

	//inputs -FeeCell
	feeCapacity := uint64(1e4)
	needCapacity := feeCapacity + keylistCfgOutput.Capacity
	liveCell, totalCapacity, err := h.dasCore.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache:          h.dasCache,
		LockScript:        h.serverScript,
		CapacityNeed:      needCapacity,
		CapacityForChange: common.MinCellOccupiedCkb,
		SearchOrder:       indexer.SearchOrderAsc,
	})

	if err != nil {
		return nil, fmt.Errorf("GetBalanceCells err: %s", err.Error())
	}
	//change cell
	if change := totalCapacity - needCapacity; change > 0 {
		splitCkb := 2000 * common.OneCkb
		if config.Cfg.Server.SplitCkb > 0 {
			splitCkb = config.Cfg.Server.SplitCkb * common.OneCkb
		}
		changeList, err := core.SplitOutputCell2(change, splitCkb, 200, h.serverScript, nil, indexer.SearchOrderAsc)
		if err != nil {
			return nil, fmt.Errorf("SplitOutputCell2 err: %s", err.Error())
		}
		for i := 0; i < len(changeList); i++ {
			txParams.Outputs = append(txParams.Outputs, changeList[i])
			txParams.OutputsData = append(txParams.OutputsData, []byte{})
		}
	}
	for _, v := range liveCell {
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.OutPoint,
		})
	}

	//cell deps
	contractDas, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	configMain, err := core.GetDasContractInfo(common.DasContractNameConfigCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	keyListCfgCell, err := core.GetDasContractInfo(common.DasKeyListConfigCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}

	actionWitness, err := witness.GenActionDataWitness(common.DasActionCreateKeyList, common.Hex2Bytes(common.ParamOwner))
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	txParams.CellDeps = append(txParams.CellDeps,
		contractDas.ToCellDep(),
		configMain.ToCellDep(),
		keyListCfgCell.ToCellDep(),
	)
	return nil, nil
}
