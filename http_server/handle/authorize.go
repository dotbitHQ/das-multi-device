package handle

import (
	"das-multi-device/cache"
	"das-multi-device/config"
	"das-multi-device/tables"
	"das-multi-device/tool"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
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
	Notes            string                     `json:"notes" binding:"max=100"`
	Avatar           int                        `json:"avatar" binding:"lt=40"`
	MasterNotes      string                     `json:"master_notes" binding:"max=100"`
}

type RespAuthorize struct {
	SignInfo
}

type reqBuildWebauthnTx struct {
	Action                common.DasAction
	Operation             common.WebAuchonKeyOperate
	ChainType             common.ChainType
	keyListConfigOutPoint string
	KeyListConfigCell     *types.CellOutput
	MasterCkbAddr         string
	MasterPayLoad         []byte
	SlavePayload          []byte
	Capacity              uint64 `json:"capacity"`
	Notes                 string `json:"notes"`
	Avatar                int    `json:"avatar"`
	MasterNotes           string `json:"master_notes"`
}

func (h *HttpHandle) Authorize(ctx *gin.Context) {
	var (
		funcName = "Authorize"
		clientIp = GetClientIp(ctx)
		req      ReqAuthorize
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		tool.Log(ctx).Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	tool.Log(ctx).Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doAuthorize(&req, &apiResp); err != nil {
		tool.Log(ctx).Error("doAuthorize err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doAuthorize(req *ReqAuthorize, apiResp *http_api.ApiResp) (err error) {
	var resp RespAuthorize
	var keyListConfigCellOutPoint string
	masterAddressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     common.ChainTypeWebauthn,
		AddressNormal: req.MasterCkbAddress,
	})
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, err.Error())
		return err
	}
	cid1 := common.Bytes2Hex(masterAddressHex.AddressPayload[:10])
	//Check if cid is enabled keyListConfigCell
	tool.Log(nil).Info("cid1: ", cid1)
	res, err := h.dbDao.GetCidPk(cid1)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "search cidpk err")
		return fmt.Errorf("SearchCidPk err: %s", err.Error())
	}
	keyListConfigCellOutPoint = res.Outpoint
	tool.Log(nil).Info("db outpoint: ", keyListConfigCellOutPoint)
	//if it is a newly created KeyListConfigCell, use it to buildWebauthnTx()
	var keyListConfigCell *types.CellOutput
	if res.Id == 0 || res.EnableAuthorize == tables.EnableAuthorizeOff {
		if req.Operation == common.DeleteWebAuthnKey { //delete from keyList
			apiResp.ApiRespErr(http_api.ApiCodeHasNoAccessToRemove, "master addr hasn`t enable authorze yet")
			return nil
		}
		canCreate, err := h.checkCanBeCreated(masterAddressHex.AddressHex)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeError500, "check if can be created err")
			return fmt.Errorf("checkCanBeCreated err : %s", err.Error())
		}
		if !canCreate {
			apiResp.ApiRespErr(http_api.ApiCodeHasNoAccessToCreate, "the main device does not have permission to activate backup")
			return fmt.Errorf("the main device does not have permission to activate backup")
		}

		//create keyListConfigCell
		keyListConfigCellOutPoint, keyListConfigCell, err = h.createKeyListCfgCell(masterAddressHex.AddressHex)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeCreateConfigCellFail, "create keyListConfigCell err")
			return fmt.Errorf("createKeyListCfgCell err: %s", err.Error())
		}
	}
	tool.Log(nil).Info("outpoint: ", keyListConfigCellOutPoint)
	//update keyListConfigCell (add das-lock-key)
	slaveAddressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     common.ChainTypeWebauthn,
		AddressNormal: req.SlaveCkbAddress,
	})
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, err.Error())
		return err
	}

	reqBuildWebauthnTx := reqBuildWebauthnTx{
		Action:                common.DasActionUpdateKeyList,
		Operation:             req.Operation,
		ChainType:             common.ChainTypeWebauthn,
		keyListConfigOutPoint: keyListConfigCellOutPoint,
		KeyListConfigCell:     keyListConfigCell,
		MasterCkbAddr:         req.MasterCkbAddress,
		MasterPayLoad:         masterAddressHex.AddressPayload,
		SlavePayload:          slaveAddressHex.AddressPayload,
		Capacity:              0,
		Avatar:                req.Avatar,
		Notes:                 req.Notes,
		MasterNotes:           req.MasterNotes,
	}

	txParams, err := h.buildUpdateAuthorizeTx(&reqBuildWebauthnTx)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "build tx err: "+err.Error())
		return fmt.Errorf("buildAddAuthorizeTx err: %s", err.Error())
	}
	if si, err := h.buildWebauthnTx(&reqBuildWebauthnTx, txParams); err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "buildWebauthnTx tx err: "+err.Error())
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

	// inputs account cell
	keyListCfgOutPoint := common.String2OutPointStruct(req.keyListConfigOutPoint)
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
	tool.Log(nil).Info("res.Transaction.Outputs: ", res.Transaction.Outputs)
	if len(res.Transaction.Outputs) == 0 {
		return nil, fmt.Errorf("KeyListCfgTransaction not exists")
	}
	txParams.Outputs = append(txParams.Outputs, res.Transaction.Outputs[0])

	builder, err := witness.WebAuthnKeyListDataBuilderFromTx(res.Transaction, common.DataTypeNew)
	if err != nil {
		return nil, fmt.Errorf("WebAuthnKeyListDataBuilderFromTx err: %s", err.Error())
	}
	var webAuthnKey witness.WebauthnKey
	webAuthnKey.MinAlgId = uint8(common.DasAlgorithmIdWebauthn)
	webAuthnKey.SubAlgId = uint8(common.DasWebauthnSubAlgorithmIdES256)
	webAuthnKey.Cid = common.Bytes2Hex(req.SlavePayload[:10])
	webAuthnKey.PubKey = common.Bytes2Hex(req.SlavePayload[10:])

	nowKeyList := witness.ConvertToWebauthnKeyList(builder.DeviceKeyListCellData.Keys())
	var newKeyList []witness.WebauthnKey
	tool.Log(nil).Info("nowKeyList: ", nowKeyList)
	tool.Log(nil).Info("slaveKey: ", webAuthnKey)
	//add webAuthnKey
	if req.Operation == common.AddWebAuthnKey {
		if len(nowKeyList) > 9 {
			return nil, fmt.Errorf("Backup devices cannot exceed 10")
		}

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
	)
	return &txParams, nil
}

func (h *HttpHandle) buildWebauthnTx(req *reqBuildWebauthnTx, txParams *txbuilder.BuildTransactionParams) (*SignInfo, error) {
	txBuilder := txbuilder.NewDasTxBuilderFromBase(h.txBuilderBase, nil)

	//if it is a newly created KeyListConfigCell, use it in req
	if req.KeyListConfigCell != nil {
		txBuilder.MapInputsCell[req.keyListConfigOutPoint] = &types.CellWithStatus{
			Cell: &types.CellInfo{
				Data:   nil,
				Output: req.KeyListConfigCell,
			},
			Status: "",
		}
	}

	if err := txBuilder.BuildTransaction(txParams); err != nil {
		return nil, fmt.Errorf("txBuilder.BuildTransaction err: %s", err.Error())
	}

	var skipGroups []int
	var sic SignInfoCache
	switch req.Action {
	case common.DasActionUpdateKeyList:
		sizeInBlock, _ := txBuilder.Transaction.SizeInBlock()
		changeCapacity := txBuilder.Transaction.Outputs[0].Capacity - sizeInBlock - 1000
		txBuilder.Transaction.Outputs[0].Capacity = changeCapacity
		tool.Log(nil).Info("buildTx:", req.Action, sizeInBlock, changeCapacity)
		if req.Operation == common.AddWebAuthnKey {
			sic.Notes = req.Notes
			sic.Avatar = req.Avatar
			sic.BackupCid = common.Bytes2Hex(req.SlavePayload[:10])
			sic.MasterNotes = req.MasterNotes
		}
	}
	signList, err := txBuilder.GenerateDigestListFromTx(skipGroups)
	if err != nil {
		return nil, fmt.Errorf("txBuilder.GenerateDigestListFromTx err: %s", err.Error())
	}

	tool.Log(nil).Info("buildTx:", txBuilder.TxString())

	sic.Action = req.Action
	sic.ChainType = req.ChainType
	sic.Address = req.MasterCkbAddr
	sic.KeyListCfgCellOpt = req.keyListConfigOutPoint
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
func (h *HttpHandle) createKeyListCfgCell(payload string) (outPoint string, outPointCell *types.CellOutput, err error) {
	delFunc, err := h.rc.LockWithRedis(common.ChainTypeWebauthn, payload, cache.CreateKeyListConfigCell, time.Minute*4)
	if err != nil {
		return "", nil, fmt.Errorf("createKeyListCfgCell LockWithRedis err :%s", err.Error())
	}
	defer func() {
		if err := delFunc(); err != nil {
			tool.Log(nil).Errorf("createKeyListCfgCell delete redis key err: %s", err)
		}
	}()

	txParams, err := h.buildCreateKeyListCfgTx(payload)
	if err != nil {
		return "", nil, err
	}

	txBuilder := txbuilder.NewDasTxBuilderFromBase(h.txBuilderBase, nil)
	if err := txBuilder.BuildTransaction(txParams); err != nil {
		return "", nil, fmt.Errorf("txBuilder.BuildTransaction err: %s", err.Error())
	}
	sizeInBlock, _ := txBuilder.Transaction.SizeInBlock()
	changeFeeIdx := len(txBuilder.Transaction.Outputs) - 1
	changeCapacity := txBuilder.Transaction.Outputs[changeFeeIdx].Capacity - sizeInBlock - 1000
	txBuilder.Transaction.Outputs[changeFeeIdx].Capacity = changeCapacity

	txHash, err := txBuilder.SendTransaction()
	if err != nil {
		return "", nil, err
	}
	outpoint := common.OutPoint2String(txHash.Hex(), 0)

	return outpoint, txParams.Outputs[0], nil
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
	if keyListCfgOutput.Capacity < 161*common.OneCkb {
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
		changeList, err := core.SplitOutputCell2(change, splitCkb, 30, h.serverScript, nil, indexer.SearchOrderAsc)
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

type AuthorizeInfo struct {
	Address string `json:"address"`
	Avatar  int    `json:"avatar"`
	Notes   string `json:"notes"`
}
type RespAuthorizeInfo struct {
	CanAuthorize int             `json:"can_authorize"`
	CkbAddress   []AuthorizeInfo `json:"ckb_address"`
}

func (h *HttpHandle) AuthorizeInfo(ctx *gin.Context) {
	var (
		funcName = "AuthorizeInfo"
		clientIp = GetClientIp(ctx)
		req      *ReqAuthorizeInfo
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		tool.Log(ctx).Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}

	tool.Log(ctx).Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doAuthorizeInfo(req, &apiResp); err != nil {
		tool.Log(ctx).Error("doIfEnableAuthorize err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doAuthorizeInfo(req *ReqAuthorizeInfo, apiResp *http_api.ApiResp) (err error) {
	var resp RespAuthorizeInfo
	masterAddressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     common.ChainTypeWebauthn,
		AddressNormal: req.CkbAddress,
	})
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, err.Error())
		return err
	}
	masterCid := common.Bytes2Hex(masterAddressHex.AddressPayload[:10])
	res, err := h.dbDao.GetCidPk(masterCid)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "Search cidpk err")
		return fmt.Errorf("SearchCidPk err: %s", err.Error())
	}
	//resp.EnableAuthorize = int(res.EnableAuthorize)
	authorizeList := make([]AuthorizeInfo, 0)
	fmt.Println("res:: ", res)
	if res.EnableAuthorize == tables.EnableAuthorizeOn {
		if res.Outpoint == "" {
			apiResp.ApiRespErr(http_api.ApiCodeError500, "outpoint is empty")
			return fmt.Errorf("outpoint is empty")
		}
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
			temp := AuthorizeInfo{}
			key := keys.Get(i)
			mId, _ := molecule.Bytes2GoU8(key.MainAlgId().RawData())
			subId, _ := molecule.Bytes2GoU8(key.SubAlgId().RawData())
			cid1 := key.Cid().AsSlice()
			pk1 := key.Pubkey().AsSlice()

			if masterAddressHex.DasSubAlgorithmId == common.DasSubAlgorithmId(subId) &&
				masterAddressHex.AddressHex == common.CalculateWebauthnPayload(cid1, pk1) {
				continue
			}

			addrNormal, err := h.dasCore.Daf().HexToNormal(core.DasAddressHex{
				DasAlgorithmId:    common.DasAlgorithmId(mId),
				DasSubAlgorithmId: common.DasSubAlgorithmId(subId),
				AddressHex:        common.CalculateWebauthnPayload(cid1, pk1),
			})
			if err != nil {
				return err
			}
			temp.Address = addrNormal.AddressNormal
			avatarNotes, err := h.dbDao.GetAvatarNotes(masterCid, common.Bytes2Hex(cid1))
			if err != nil {
				apiResp.ApiRespErr(http_api.ApiCodeDbError, "GetAvatarNotes err")
				return fmt.Errorf("GetAvatarNotes err: %s", err.Error())
			}
			if avatarNotes.Id != 0 {
				temp.Notes = avatarNotes.Notes
				temp.Avatar = avatarNotes.Avatar
			}
			authorizeList = append(authorizeList, temp)

		}

	}
	resp.CkbAddress = authorizeList
	if res.EnableAuthorize == 0 {
		canCreate, err := h.checkCanBeCreated(masterAddressHex.AddressHex)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeError500, "check if can be created err")
			return fmt.Errorf("checkCanBeCreated err : %s", err.Error())
		}
		if canCreate {
			resp.CanAuthorize = 1
		}
	} else {
		resp.CanAuthorize = 1
	}

	apiResp.ApiRespOK(resp)
	return nil
}
