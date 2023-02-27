package inited

import "gen-piece-commitment/config"

func InitApp() {
	InitDB(config.Cfg.MySQL.Host, config.Cfg.MySQL.User, config.Cfg.MySQL.Password, config.Cfg.MySQL.DbName)
	InitRedis(config.Cfg.Redis.Host, config.Cfg.Redis.Password, config.Cfg.Redis.Db, config.Cfg.Redis.PoolSize)
}
