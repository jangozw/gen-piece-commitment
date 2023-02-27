package config

import (
	"fmt"
	"gen-piece-commitment/util"
	"github.com/BurntSushi/toml"
)

var Cfg *Config

type LogLevel struct {
	Name  string `json:"name"`
	Level string `json:"level"`
}

type Config struct {
	Redis struct {
		Host     string `json:"host"`
		Password string `json:"password"`
		Db       int    `json:"db"`
		PoolSize int    `json:"poolsize"`
	} `json:"redis"`

	MySQL struct {
		Host     string `json:"host"`
		User     string `json:"user"`
		Password string `json:"password"`
		DbName   string `json:"dbname"`
	} `json:"mysql"`

	Log struct {
		Levels []LogLevel `json:"level"`
	} `json:"log"`
}

func InitConfig(cf string) {
	if Cfg != nil {
		return
	}
	if ok, _ := util.IsPathExists(cf); !ok {
		panic(fmt.Sprintf("%s is not exists", cf))
	}
	cg := &Config{}
	_, err := toml.DecodeFile(cf, cg)
	if err != nil {
		panic(err)
	}
	Cfg = cg
	fmt.Printf("config: %+v\n", Cfg)
	return
}

func GetLogLevel(name string) string {
	m := make(map[string]string)
	for _, v := range Cfg.Log.Levels {
		m[v.Name] = v.Level
	}
	if v, ok := m[name]; ok {
		return v
	}
	return "debug"
}
