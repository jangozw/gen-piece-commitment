package handler

import (
	"gen-piece-commitment/config"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("handler")

func InitLog() {
	logging.SetLogLevel("handler", config.GetLogLevel("handler"))
}
