package handle

import (
	"das-multi-device/cache"
	"das-multi-device/config"
	"das-multi-device/http_server/api_code"
	"das-multi-device/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/molecule"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"time"
)

type ReqAuthorize struct {
	MasterCkbAddress string                     `json:"master_ckb_address" binding:"required"`
	SlaveCkbAddress  string                     `json:"slave_ckb_address" binding:"required"`
	Operation        common.WebAuchonKeyOperate `json:"operation" binding:"required"` //operation = 0 删除，1 添加
}

type RespAuthorize struct {
	SignInfo
}

type reqBuildWebauthnTx struct {
	Action          common.DasAction
	Operation       common.WebAuchonKeyOperate
	ChainType       common.ChainType
	keyListConfigOp string
	MasterPayLoad   []byte
	SlavePayload    []byte
	Capacity        uint64 `json:"capacity"`
}

func (h *HttpHandle) Authorize(ctx *gin.Context) {
	var (
		funcName = "Authorize"
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
	masterAddressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     common.ChainTypeWebauthn,
		AddressNormal: req.MasterCkbAddress,
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return err
	}
	masterPayloadHex := common.Bytes2Hex(masterAddressHex.AddressPayload)
	cid1 := common.Bytes2Hex(masterAddressHex.AddressPayload[:10])
	//Check if cid is enabled keyListConfigCell
	res, err := h.dbDao.GetCidPk(cid1)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search cidpk err")
		return fmt.Errorf("SearchCidPk err: %s", err.Error())
	}
	keyListConfigCellOutPoint = res.Outpoint

	if res.Id == 0 || res.EnableAuthorize == tables.EnableAuthorizeOff {
		if req.Operation == common.DeleteWebAuthnKey { //delete from keyList
			apiResp.ApiRespErr(api_code.ApiCodeHasNoAccessToRemove, "master addr hasn`t enable authorze yet")
			return fmt.Errorf("SearchCidPk err: %s", err.Error())
		}
		//Check if keyListConfigCell can be created
		if config.Cfg.Server.Net == common.DasNetTypeMainNet {
			canCreate, err := h.checkCanBeCreated(masterPayloadHex)
			if err != nil {
				apiResp.ApiRespErr(api_code.ApiCodeError500, "check if can be created err")
				return fmt.Errorf("checkCanBeCreated err : %s", err.Error())
			}
			if !canCreate {
				apiResp.ApiRespErr(api_code.ApiCodeHasNoAccessToCreate, "master_address has no access to enable authorize")
				return fmt.Errorf("master_address hasn`t enable authorize")
			}
		}

		//create keyListConfigCell
		keyListConfigCellOutPoint, err = h.createKeyListCfgCell(masterPayloadHex)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeCreateConfigCellFail, "create keyListConfigCell err")
			return err
		}
	}

	//update keyListConfigCell (add das-lock-key)
	slaveAddressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     common.ChainTypeWebauthn,
		AddressNormal: req.SlaveCkbAddress,
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return err
	}

	reqBuildWebauthnTx := reqBuildWebauthnTx{
		Action:          common.DasActionUpdateKeyList,
		Operation:       req.Operation,
		ChainType:       common.ChainTypeWebauthn,
		keyListConfigOp: keyListConfigCellOutPoint,
		MasterPayLoad:   masterAddressHex.AddressPayload,
		SlavePayload:    slaveAddressHex.AddressPayload,
		Capacity:        0,
	}

	txParams, err := h.buildUpdateAuthorizeTx(&reqBuildWebauthnTx)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "build tx err: "+err.Error())
		return fmt.Errorf("buildAddAuthorizeTx err: %s", err.Error())
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

func (h *HttpHandle) buildUpdateAuthorizeTx(req *reqBuildWebauthnTx) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams
	contractDas, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	configMain, err := core.GetDasContractInfo(common.DasContractNameConfigCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	keyListCfgCell, err := core.GetDasContractInfo(common.DasKeyListCellType)
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

	res, err := h.dasCore.Client().GetTransaction(h.ctx, keyListCfgOutPoint.TxHash)
	if err != nil {
		return nil, fmt.Errorf("GetTransaction err: %s", err.Error())
	}
	//capacity := res.Transaction.Outputs[keyListCfgOutPoint.Index].Capacity

	txParams.Outputs = append(txParams.Outputs, res.Transaction.Outputs[0])

	builder, err := witness.WebAuthnKeyListDataBuilderFromTx(res.Transaction, common.DataTypeNew)
	if err != nil {
		return nil, fmt.Errorf("WebAuthnKeyListDataBuilderFromTx err: %s", err.Error())
	}
	var webAuthnKey witness.WebauthnKey
	webAuthnKey.MinAlgId = uint8(common.DasAlgorithmIdWebauthn)
	webAuthnKey.SubAlgId = uint8(common.DasWebauthnSubAlgorithmIdES256)
	webAuthnKey.Cid = string(req.SlavePayload[:10])
	webAuthnKey.PubKey = string(req.SlavePayload[10:])

	nowKeyList := witness.ConvertToWebauthnKeyList(builder.DeviceKeyListCellData.Keys())
	var newKeyList []witness.WebauthnKey
	//add webAuthnKey
	if req.Operation == common.AddWebAuthnKey {
		for _, v := range nowKeyList {
			if v.Cid == webAuthnKey.Cid && v.PubKey == webAuthnKey.PubKey {
				return nil, fmt.Errorf("Cannot add repeatedly")
			}
		}
		nowKeyList = append(nowKeyList, webAuthnKey)
		newKeyList = nowKeyList
	} else { //delete webAuthnKey
		isExist := false
		for _, v := range nowKeyList {
			if v.Cid == webAuthnKey.Cid && v.PubKey == webAuthnKey.PubKey {
				isExist = true
			} else {
				newKeyList = append(newKeyList, v)
			}
		}
		if !isExist {
			return nil, fmt.Errorf("this deviceKey isn`t exist")
		}
	}

	klWitness, klData, err := builder.GenWitness(&witness.WebauchnKeyListCellParam{
		Action:            common.DasActionUpdateKeyList,
		Operation:         req.Operation,
		OldIndex:          0,
		NewIndex:          0,
		UpdateWebauthnKey: newKeyList,
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
	case common.DasActionUpdateKeyList:
		//TODO 计算手续费
		sizeInBlock, _ := txBuilder.Transaction.SizeInBlock()
		changeCapacity := txBuilder.Transaction.Outputs[0].Capacity - sizeInBlock - 1000
		txBuilder.Transaction.Outputs[0].Capacity = changeCapacity
		log.Info("buildTx:", req.Action, sizeInBlock, changeCapacity)
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
	//check if u have .bit account
	num, err := h.dbDao.GetAccountInfos(payload)
	if err != nil {
		return false, fmt.Errorf("GetAccountInfos err: %s", err.Error())
	}
	if num > 0 {
		return true, nil
	}

	//check if you have ckb amount
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

// create keyListConfigCell
func (h *HttpHandle) createKeyListCfgCell(payload string) (outPoint string, err error) {
	delFunc, err := h.rc.LockWithRedis(common.ChainTypeWebauthn, payload, cache.CreateKeyListConfigCell, time.Minute*4)
	if err != nil {
		return "", fmt.Errorf("createKeyListCfgCell LockWithRedis err :%s", err.Error())
	}
	defer func() {
		if err := delFunc(); err != nil {
			log.Errorf("createKeyListCfgCell delete redis key err: %s", err)
		}
	}()

	txParams, err := h.buildCreateKeyListCfgTx(payload)
	if err != nil {
		return "", err
	}

	txBuilder := txbuilder.NewDasTxBuilderFromBase(h.txBuilderBase, nil)
	if err := txBuilder.BuildTransaction(txParams); err != nil {
		return "", fmt.Errorf("txBuilder.BuildTransaction err: %s", err.Error())
	}
	sizeInBlock, _ := txBuilder.Transaction.SizeInBlock()
	changeFeeIdx := len(txBuilder.Transaction.Outputs) - 1
	changeCapacity := txBuilder.Transaction.Outputs[changeFeeIdx].Capacity - sizeInBlock - 1000
	txBuilder.Transaction.Outputs[changeFeeIdx].Capacity = changeCapacity

	txHash, err := txBuilder.SendTransaction()
	if err != nil {
		return "", err
	}
	outpoint := common.OutPoint2String(txHash.Hex(), 0)

	return outpoint, nil
}

func (h *HttpHandle) buildCreateKeyListCfgTx(webauthnPayload string) (*txbuilder.BuildTransactionParams, error) {
	txParams := &txbuilder.BuildTransactionParams{}
	//cell deps
	contractDas, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	configMain, err := core.GetDasContractInfo(common.DasContractNameConfigCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	keyListCfgCell, err := core.GetDasContractInfo(common.DasKeyListCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	txParams.CellDeps = append(txParams.CellDeps,
		contractDas.ToCellDep(),
		configMain.ToCellDep(),
		keyListCfgCell.ToCellDep(),
	)

	// OutputData
	payloadBytes := common.Hex2Bytes(webauthnPayload)
	cid1 := payloadBytes[:10]
	cid1Byte10, err := molecule.GoBytes2MoleculeByte10(cid1)
	if err != nil {
		return nil, err
	}
	pk1 := payloadBytes[10:]
	pk1Byte10, err := molecule.GoBytes2MoleculeByte10(pk1)
	if err != nil {
		return nil, err
	}
	deviceKeyBuilder := molecule.NewDeviceKeyBuilder()
	deviceKeyBuilder.MainAlgId(molecule.GoU8ToMoleculeU8(uint8(common.ChainTypeWebauthn)))
	deviceKeyBuilder.SubAlgId(molecule.GoU8ToMoleculeU8(uint8(common.DasWebauthnSubAlgorithmIdES256)))
	deviceKeyBuilder.Cid(cid1Byte10)
	deviceKeyBuilder.Pubkey(pk1Byte10)
	keyListBuilder := molecule.NewDeviceKeyListBuilder()
	keyListBuilder.Push(deviceKeyBuilder.Build())
	deviceKeyList := keyListBuilder.Build()

	scriptBuilder := molecule.NewScriptBuilder()
	scriptBuilder.Args(molecule.GoBytes2MoleculeBytes(h.serverScript.Args))
	codeHash, err := molecule.HashFromSlice(h.serverScript.CodeHash.Bytes(), true)
	if err != nil {
		return nil, err
	}
	scriptBuilder.CodeHash(*codeHash)
	hashType, err := h.serverScript.HashType.Serialize()
	if err != nil {
		return nil, err
	}
	scriptBuilder.HashType(molecule.NewByte(hashType[0]))

	cellDataBuilder := molecule.NewDeviceKeyListCellDataBuilder()
	cellDataBuilder.Keys(deviceKeyList)
	cellDataBuilder.RefundLock(scriptBuilder.Build())
	cellData := cellDataBuilder.Build()

	webAuthnBuilder := witness.WebAuthnKeyListDataBuilder{}
	webAuthnBuilder.DeviceKeyListCellData = &cellData
	webAuthnBuilder.Version = common.GoDataEntityVersion3

	klWitness, klData, err := webAuthnBuilder.GenWitness(&witness.WebauchnKeyListCellParam{
		Action: common.DasActionCreateKeyList,
	})
	txParams.OutputsData = append(txParams.OutputsData, klData)

	ownerHex := core.DasAddressHex{
		DasAlgorithmId: common.DasAlgorithmIdWebauthn,
		AddressHex:     webauthnPayload,
		IsMulti:        false,
		ChainType:      common.ChainTypeWebauthn,
	}
	lockArgs, err := h.dasCore.Daf().HexToArgs(ownerHex, ownerHex)
	if err != nil {
		return nil, fmt.Errorf("HexToArgs err: %s", err.Error())
	}

	//outputs
	keyListCfgOutput := &types.CellOutput{
		Lock: contractDas.ToScript(lockArgs),
		Type: keyListCfgCell.ToScript(nil),
	}
	keyListCfgOutput.Capacity = keyListCfgOutput.OccupiedCapacity(klData) * common.OneCkb
	if keyListCfgOutput.Capacity < 161 * common.OneCkb {
		keyListCfgOutput.Capacity = 161 * common.OneCkb
	}
	txParams.Outputs = append(txParams.Outputs, keyListCfgOutput)

	//inputs -FeeCell
	feeCapacity := uint64(1e4)
	needCapacity := feeCapacity + keyListCfgOutput.Capacity
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
	for _, v := range liveCell {
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.OutPoint,
		})
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

	//Witness
	actionWitness, err := witness.GenActionDataWitness(common.DasActionCreateKeyList, common.Hex2Bytes(common.ParamOwner))
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)
	txParams.Witnesses = append(txParams.Witnesses, klWitness)

	return txParams, nil
}

type ReqAuthorizeInfo struct {
	CkbAddress string `json:"ckb_address" binding:"required"`
}
type RespAuthorizeInfo struct {
	EnableAuthorize int      `json:"enable_authorize" binding:"required"`
	CkbAddress      []string `json:"ckb_address"`
}

func (h *HttpHandle) AuthorizeInfo(ctx *gin.Context) {
	var (
		funcName = "IfEnableAuthorize"
		clientIp = GetClientIp(ctx)
		req      *ReqAuthorizeInfo
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

	if err = h.doAuthorizeInfo(req, &apiResp); err != nil {
		log.Error("doIfEnableAuthorize err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doAuthorizeInfo(req *ReqAuthorizeInfo, apiResp *api_code.ApiResp) (err error) {
	var resp RespAuthorizeInfo
	masterAddressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     common.ChainTypeWebauthn,
		AddressNormal: req.CkbAddress,
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return err
	}
	cid1 := common.Bytes2Hex(masterAddressHex.AddressPayload[:10])
	res, err := h.dbDao.GetCidPk(cid1)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "Search cidpk err")
		return fmt.Errorf("SearchCidPk err: %s", err.Error())
	}
	resp.EnableAuthorize = int(res.EnableAuthorize)
	resp.CkbAddress = make([]string, 0)

	if res.EnableAuthorize == tables.EnableAuthorizeOn {
		outpoint := common.String2OutPointStruct(res.Outpoint)
		tx, err := h.dasCore.Client().GetTransaction(h.ctx, outpoint.TxHash)
		if err != nil {
			return err
		}
		builder, err := witness.WebAuthnKeyListDataBuilderFromTx(tx.Transaction, common.DataTypeNew)
		if err != nil {
			return err
		}
		keys := builder.DeviceKeyListCellData.Keys()
		for i := uint(0); i < keys.Len(); i++ {
			key := keys.Get(i)
			mId, _ := molecule.Bytes2GoU8(key.MainAlgId().RawData())
			subId, _ := molecule.Bytes2GoU8(key.SubAlgId().RawData())
			cid1 := key.Cid().AsSlice()
			pk1 := key.Pubkey().AsSlice()

			if masterAddressHex.DasSubAlgorithmId == common.DasSubAlgorithmId(subId) &&
				masterAddressHex.AddressHex == common.Bytes2Hex(append(cid1, pk1...)) {
				continue
			}

			addrNormal, err := h.dasCore.Daf().HexToNormal(core.DasAddressHex{
				DasAlgorithmId:    common.DasAlgorithmId(mId),
				DasSubAlgorithmId: common.DasSubAlgorithmId(subId),
				AddressHex:        common.CaculateWebauthnPayload(cid1, pk1),
			})
			if err != nil {
				return err
			}
			resp.CkbAddress = append(resp.CkbAddress, addrNormal.AddressNormal)
		}
	}
	apiResp.ApiRespOK(resp)
	return nil
}
