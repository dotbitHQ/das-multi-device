package block_parser

import (
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
)

func (b *BlockParser) ActionUpdateKeylist(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasKeyListCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		return
	}
	outpoint := common.OutPoint2String(req.TxHash, 0)
	if err := b.DbDao.UpdatePendingStatusToConfirm(req.Action, outpoint, req.BlockNumber, req.BlockTimestamp); err != nil {
		resp.Err = fmt.Errorf("UpdatePendingStatusToConfirm err: %s", err.Error())
		return
	}
	if err := b.DbDao.UpdateAvatrNotesToConfirm(outpoint, req.BlockNumber, req.BlockTimestamp); err != nil {
		resp.Err = fmt.Errorf("UpdateAvatrNotesToConfirm err: %s", err.Error())
		return
	}
	return
}
