package tables

import (
	"github.com/dotbitHQ/das-lib/common"
	"time"
)

type TableAuthorize struct {
	Id             uint64                `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	MasterAlgId    common.DasAlgorithmId `json:"master_alg_id" gorm:"column:master_alg_id; type:tinyint DEFAULT NULL"`
	MasterSubAlgId common.DasAlgorithmId `json:"master_sub_alg_id" gorm:"column:master_sub_alg_id; type:tinyint DEFAULT NULL"`
	MasterCid      string                `json:"master_cid" gorm:"column:master_cid; uniqueIndex:uk_mastercid_slavecid;  type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL"`
	MasterPk       string                `json:"master_pk" gorm:"column:master_pk; type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL"`
	SlaveAlgId     common.DasAlgorithmId `json:"slave_alg_id" gorm:"column:slave_alg_id; type:tinyint DEFAULT NULL"`
	SlaveSubAlgId  common.DasAlgorithmId `json:"slave_sub_alg_id" gorm:"column:slave_sub_alg_id; type:tinyint DEFAULT NULL"`
	SlaveCid       string                `json:"slave_cid" gorm:"column:slave_cid; uniqueIndex:uk_mastercid_slavecid;  type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL"`
	SlavePk        string                `json:"slave_pk" gorm:"column:slave_pk; type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL"`
	Outpoint       string                `json:"outpoint" gorm:"column:outpoint;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL"`
	CreatedAt      time.Time             `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''""`
	UpdatedAt      time.Time             `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const (
	TableNameAuthorize = "t_authorize"
)

func (t *TableAuthorize) TableName() string {
	return TableNameAuthorize
}
