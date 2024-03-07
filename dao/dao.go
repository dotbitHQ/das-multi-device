package dao

import (
	"das-multi-device/config"
	"das-multi-device/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/http_api"
	"gorm.io/gorm"
)

type DbDao struct {
	db       *gorm.DB
	parserDb *gorm.DB
}

func NewGormDB(dbMysql, parserMysql config.DbMysql, autoMigrate bool) (*DbDao, error) {
	db, err := http_api.NewGormDB(dbMysql.Addr, dbMysql.User, dbMysql.Password, dbMysql.DbName, dbMysql.MaxOpenConn, dbMysql.MaxIdleConn)
	if err != nil {
		return nil, fmt.Errorf("toolib.NewGormDB err: %s", err.Error())
	}

	// AutoMigrate will create tables, missing foreign keys, constraints, columns and indexes.
	// It will change existing column’s type if its size, precision, nullable changed.
	// It WON’T delete unused columns to protect your data.

	parserDb, err := http_api.NewGormDB(parserMysql.Addr, parserMysql.User, parserMysql.Password, parserMysql.DbName, parserMysql.MaxOpenConn, parserMysql.MaxIdleConn)
	if err != nil {
		return nil, fmt.Errorf("toolib.NewGormDB err: %s", err.Error())
	}
	if autoMigrate {
		if err = db.AutoMigrate(
			&tables.TableWebauthnPendingInfo{},
			&tables.TableBlockParserInfo{},
			&tables.TableAvatarNotes{},
			&tables.TableCidInfo{},
			&tables.TableCidPk{},
		); err != nil {
			return nil, err
		}
	}

	return &DbDao{db: db, parserDb: parserDb}, nil
}
