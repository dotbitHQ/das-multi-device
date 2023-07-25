package dao

import (
	"das-multi-device/tables"
	"fmt"
)

func (d *DbDao) GetMasters(cid1 string) (authorize []*tables.TableAuthorize, err error) {
	fmt.Println("cid1: ", cid1)
	err = d.parserDb.Where("`slave_cid`=?", cid1).Find(&authorize).Error
	return
}
