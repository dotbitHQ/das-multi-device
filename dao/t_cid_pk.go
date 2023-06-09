package dao

import "das-multi-device/tables"

func (d *DbDao) GetCidPk(cid1 string) (authorize tables.TableCidPk, err error) {
	err = d.parserDb.Where("`cid`= ? ", cid1).Find(&authorize).Error
	return
}
