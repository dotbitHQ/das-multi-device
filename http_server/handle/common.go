package handle

import (
	"crypto/md5"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"time"
)

type Pagination struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

func (p Pagination) GetLimit() int {
	if p.Size < 1 || p.Size > 100 {
		return 100
	}
	return p.Size
}

func (p Pagination) GetOffset() int {
	page := p.Page
	if p.Page < 1 {
		page = 1
	}
	size := p.GetLimit()
	return (page - 1) * size
}

type SignInfo struct {
	SignKey     string               `json:"sign_key"`               // sign tx key
	SignAddress string               `json:"sign_address,omitempty"` // sign address
	SignList    []txbuilder.SignData `json:"sign_list"`              // sign list
}

type SignInfoList struct {
	Action    common.DasAction `json:"action"`
	SubAction common.SubAction `json:"sub_action"`
	SignKey   string           `json:"sign_key"`
	List      []SignInfo       `json:"list"`
}

type SignInfoCache struct {
	ChainType         common.ChainType                   `json:"chain_type"`
	Address           string                             `json:"address"`
	Action            string                             `json:"action"`
	Capacity          uint64                             `json:"capacity"`
	KeyListCfgCellOpt string                             `json:"key_list_cfg_cell_opt"`
	BuilderTx         *txbuilder.DasTxBuilderTransaction `json:"builder_tx"`
	Avatar            int                                `json:"avatar"`
	Notes             string                             `json:"notes"`
	BackupCid         string                             `json:"backup_cid"`
	MasterNotes       string                             `json:"master_notes" `
}

func (s *SignInfoCache) SignKey() string {
	key := fmt.Sprintf("%d%s%s%d", s.ChainType, s.Address, s.Action, time.Now().UnixNano())
	return fmt.Sprintf("%x", md5.Sum([]byte(key)))
}

type SignInfoCacheList struct {
	Action        string                               `json:"action"`
	Account       string                               `json:"account"`
	TaskIdList    []string                             `json:"task_id_list"`
	BuilderTxList []*txbuilder.DasTxBuilderTransaction `json:"builder_tx_list"`
}

func (s *SignInfoCacheList) SignKey() string {
	key := fmt.Sprintf("%s%s%d", s.Account, s.Action, time.Now().UnixNano())
	return fmt.Sprintf("%x", md5.Sum([]byte(key)))
}
