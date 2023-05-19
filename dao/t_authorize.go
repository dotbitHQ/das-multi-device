package dao

import "das-multi-device/tables"

func (d *DbDao) GetMasters(cid1 string) (authorize []*tables.TableAuthorize, err error) {
	err = d.parserDb.Where("`slave_cid`=?", cid1).Find(&authorize).Error
	return
}
