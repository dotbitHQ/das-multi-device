package block_parser

import (
	"das-multi-device/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api/logger"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

var log = logger.NewLogger("block_parser", logger.LevelDebug)

func (b *BlockParser) registerTransactionHandle() {
	b.mapTransactionHandle = make(map[string]FuncTransactionHandle)
	b.mapTransactionHandle[common.DasActionUpdateKeyList] = b.ActionUpdateKeylist

}

func isCurrentVersionTx(tx *types.Transaction, name common.DasContractName) (bool, error) {
	contract, err := core.GetDasContractInfo(name)
	if err != nil {
		return false, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	isCV := false
	for _, v := range tx.Outputs {
		if v.Type == nil {
			continue
		}
		if contract.IsSameTypeId(v.Type.CodeHash) {
			isCV = true
			break
		}
	}
	return isCV, nil
}

type FuncTransactionHandleReq struct {
	DbDao          *dao.DbDao
	Tx             *types.Transaction
	TxHash         string
	BlockNumber    uint64
	BlockTimestamp uint64
	Action         common.DasAction
}

type FuncTransactionHandleResp struct {
	ActionName string
	Err        error
}

type FuncTransactionHandle func(FuncTransactionHandleReq) FuncTransactionHandleResp
