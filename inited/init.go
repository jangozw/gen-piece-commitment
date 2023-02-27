package inited

import (
	"gen-piece-commitment/config"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("inited")

func InitLog() {
	logging.SetLogLevel("inited", config.GetLogLevel("inited"))
}
