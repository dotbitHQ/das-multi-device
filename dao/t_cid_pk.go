package dao

import (
	"das-multi-device/tables"
	"gorm.io/gorm/clause"
)

func (d *DbDao) GetCidPk(cid1 string) (cidPk tables.TableCidPk, err error) {
	//cid pk from ckb chain
	err = d.parserDb.Where("`cid`= ? ", cid1).Find(&cidPk).Error
	//Centralized cid pk
	if cidPk.Id == 0 {
		err = d.db.Where("`cid`= ? ", cid1).Find(&cidPk).Error
	}
	return
}

func (d *DbDao) InsertCidPk(data tables.TableCidPk) (err error) {
	if err := d.db.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{
			"origin_pk",
		}),
	}).Create(&data).Error; err != nil {
		return err
	}
	return
}
