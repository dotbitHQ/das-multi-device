package tables

import (
	"time"
)

const (
	TableNameCidPk = "t_cid_pk"
)

type IsEnableAuthorize = uint8

const (
	EnableAuthorizeOff IsEnableAuthorize = 0
	EnableAuthorizeOn  IsEnableAuthorize = 1
)

type TableCidPk struct {
	Id              uint64            `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	Cid             string            `json:"cid" gorm:"column:cid; type:varchar(255) NOT NULL DEFAULT '0'"`
	Pk              string            `json:"pk" gorm:"column:pk; type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '';"`
	EnableAuthorize IsEnableAuthorize `json:"enable_authorize" gorm:"column:enable_authorize; type:tinyint NOT NULL DEFAULT '0'"`
	KeylistOutpoint string            `json:"keylist_outpoint" gorm:"column:outpoint;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL"`
	CreatedAt       time.Time         `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''""`
	UpdatedAt       time.Time         `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

func (t *TableCidPk) TableName() string {
	return TableNameCidPk
}
