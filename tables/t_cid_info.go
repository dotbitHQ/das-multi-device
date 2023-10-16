package tables

import "time"

type TableCidInfo struct {
	Id          uint64    `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	OriginalCid string    `json:"original_cid" gorm:"column:original_cid; uniqueIndex:uk_oringinal_cid;  type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL"`
	Cid         string    `json:"cid" gorm:"column:cid; uniqueIndex:uk_cid;  type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL"`
	Notes       string    `json:"notes" gorm:"column:notes;  type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL"`
	Device      string    `json:"device" gorm:"column:device;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''""`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const (
	TableNameCidInfo = "t_cid_info"
)

func (t *TableCidInfo) TableName() string {
	return TableNameCidInfo
}
