package dao

import "das-multi-device/tables"

func (d *DbDao) GetCidPk(cid1 string) (cidPk tables.TableCidPk, err error) {
	err = d.parserDb.Where("`cid`= ? ", cid1).Find(&cidPk).Error
	return
}
