package dao

import (
	"das-multi-device/tables"
)

func (d *DbDao) CreateCidInfo(cidInfo tables.TableCidInfo) error {
	return d.db.Create(&cidInfo).Error
}
func (d *DbDao) GetCidInfo(cid string) (cidInfo tables.TableCidInfo, err error) {
	err = d.db.Where("`cid`= ? ", cid).Find(&cidInfo).Error
	return
}
