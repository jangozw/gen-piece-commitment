package inited

import (
	"fmt"
	"gen-piece-commitment/model"
	_ "github.com/go-sql-driver/mysql" // 这个不能删
	logging "github.com/ipfs/go-log/v2"
	"github.com/jinzhu/gorm"
	"time"
)

type MysqlLog struct {
	logger *logging.ZapEventLogger
}

var DB *gorm.DB

func InitDB(host, user, pwd, dbName string) {
	if DB != nil {
		return
	}
	var mysqlLog = &MysqlLog{
		logger: log,
	}
	args := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=%s", user, pwd, host, dbName, "Local")
	db, err := gorm.Open("mysql", args)
	if err != nil {
		panic(fmt.Sprintf("couldn't connect to mysql, check your connect args in config.toml, errM: %s", err.Error()))
	}
	// config gorm db
	// 全局设置表名不可以为复数形式。
	db.SingularTable(true)
	// prevent no where cause update/delete
	db.BlockGlobalUpdate(true)
	// enable log
	db.LogMode(true)
	db.DB().SetMaxIdleConns(20)
	db.DB().SetMaxOpenConns(100)
	db.DB().SetConnMaxLifetime(3600 * time.Second)

	db.Callback().Create().Replace("gorm:update_time_stamp", func(scope *gorm.Scope) {
		scope.SetColumn("CreatedAt", time.Now())
		scope.SetColumn("UpdatedAt", time.Now())
	})
	db.Callback().Update().Replace("gorm:update_time_stamp", func(scope *gorm.Scope) {
		scope.SetColumn("UpdatedAt", time.Now())
	})
	// sql 写入日志 或控制台， 二选一
	db.SetLogger(mysqlLog)
	db.CreateTable(&model.TDealInfo{})
	db.Set("gorm:table_options", "ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci").AutoMigrate(&model.TDealInfo{})
	DB = db
	return
}

func (d *MysqlLog) Print(v ...interface{}) {
	switch v[0] {
	case "sql":
		d.logger.Debugf("%s  %s %v %s", v[2], v[3], v[4], v[1])
	case "log":
		d.logger.Debugw("sql-log", v[2])
	}
}
