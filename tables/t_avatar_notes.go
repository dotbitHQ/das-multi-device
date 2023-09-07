package tables

import "time"

type TableAvatarNotes struct {
	Id             uint64    `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	MasterCid      string    `json:"master_cid" gorm:"column:master_cid; uniqueIndex:uk_mastercid_slavecid;  type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL"`
	SlaveCid       string    `json:"slave_cid" gorm:"column:slave_cid; uniqueIndex:uk_mastercid_slavecid;  type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL"`
	Avatar         int       `json:"avatar" gorm:"column:avatar;type:smallint(6) NOT NULL DEFAULT '0'"`
	Notes          string    `json:"notes" gorm:"column:notes;  type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL"`
	BlockNumber    uint64    `json:"block_number" gorm:"column:block_number;index:k_block_number;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	BlockTimestamp uint64    `json:"block_timestamp" gorm:"column:block_timestamp;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	Outpoint       string    `json:"outpoint" gorm:"column:outpoint;uniqueIndex:uk_a_o;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	Status         int       `json:"status" gorm:"column:status;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '0-default -1-rejected 1-confirm'"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''""`
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const (
	TableNameAvatarNotes = "t_avatar_notes"
)

func (t *TableAvatarNotes) TableName() string {
	return TableNameAvatarNotes
}
