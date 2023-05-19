package dao

import "das-multi-device/tables"

//func (d *DbDao) GetAccountInfos(limit int) (accountInfos []*tables.TableAccountInfo, err error) {
//	err = d.parserDb.Select("account_id,next_account_id").Where("`parent_account_id`=''").
//		Order("account_id").Limit(limit).Find(&accountInfos).Error
//	return
//}

func (d *DbDao) GetAccountInfos(payload string) (num int64, err error) {
	err = d.parserDb.Model(tables.TableAccountInfo{}).Where("owner = ? or manager = ?", payload, payload).Count(&num).Error
	return
}
