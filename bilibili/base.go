package bilibili

import (
	"encoding/binary"
	"github.com/nutsdb/nutsdb"
	"path"
	"videoDynamicAcquisition/baseStruct"
)

var (
	db *nutsdb.DB
)

const (
	nutsdbPath = "bilidynamic.db"
	bucket     = "spiderData"
)

func initNutsdb() {
	var err error
	db, err = nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(path.Join(baseStruct.RootPath, nutsdbPath)), // 数据库会自动创建这个目录文件
	)
	if err != nil {
		panic(err)
	}
}

func saveDynamicBaseline(userId int64, baseline int64) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	var (
		keyUserId []byte
		val       []byte
	)
	binary.BigEndian.PutUint64(keyUserId, uint64(userId))
	binary.BigEndian.PutUint64(val, uint64(baseline))
	err = tx.Put(bucket, keyUserId, val, nutsdb.Persistent)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
func getDynamicBaseline(userId int64) int64 {
	return 0
}
