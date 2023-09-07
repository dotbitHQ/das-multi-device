package dao

import (
	"das-multi-device/tables"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (d *DbDao) CreateAvatarNotes(avatarNotes *tables.TableAvatarNotes) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"avatar",
				"notes",
				"outpoint",
			}),
		}).Create(&avatarNotes).Error; err != nil {
			return err
		}
		return nil
	})

	return d.db.Create(&avatarNotes).Error
}

func (d *DbDao) UpdateAvatrNotesToConfirm(outpoint string, blockNumber, blockTimestamp uint64) error {
	return d.db.Model(tables.TableAvatarNotes{}).
		Where("outpoint=?", outpoint).
		Updates(map[string]interface{}{
			"block_number":    blockNumber,
			"block_timestamp": blockTimestamp,
			"status":          tables.StatusConfirm,
		}).Error
}

func (d *DbDao) GetAvatarNotes(masterCid, slaveCid string) (avatarNotes tables.TableAvatarNotes, err error) {
	err = d.db.Where("`master_cid`= ? and `slave_cid`= ? and status = 1", masterCid, slaveCid).Find(&avatarNotes).Error
	return
}
