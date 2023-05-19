package tables

import (
	"github.com/dotbitHQ/das-lib/common"
	"time"
)

type TableAuthorize struct {
	Id             uint64                `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	MasterAlgId    common.DasAlgorithmId `json:"master_alg_id" gorm:"column:master_alg_id"`
	MasterSubAlgId common.DasAlgorithmId `json:"master_sub_alg_id" gorm:"column:master_sub_alg_id"`
	MasterCid      string                `json:"master_cid" gorm:"column:master_cid"`
	MasterPk       string                `json:"master_pk" gorm:"column:master_pk"`
	SlaveAlgId     common.DasAlgorithmId `json:"slave_alg_id" gorm:"column:slave_alg_id"`
	SlaveSubAlgId  common.DasAlgorithmId `json:"slave_sub_alg_id" gorm:"column:slave_sub_alg_id"`
	SlaveCid       string                `json:"slave_cid" gorm:"column:slave_cid"`
	SlavePk        string                `json:"slave_pk" gorm:"column:slave_pk"`
	CreatedAt      time.Time             `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time             `json:"updated_at" gorm:"column:updated_at"`
	Outpoint       string                `json:"outpoint" gorm:"column:outpoint"`
}
