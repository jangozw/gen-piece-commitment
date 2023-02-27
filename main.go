package main

import (
	"fmt"
	"gen-piece-commitment/config"
	"gen-piece-commitment/handler"
	"gen-piece-commitment/inited"
	"github.com/urfave/cli/v2"
	"os"
)

var flags = []cli.Flag{
	&cli.StringFlag{
		Name:     "config",
		Aliases:  []string{"c"},
		Value:    "config.toml",
		Required: true,
		Usage:    "--config=config.toml",
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "gen-piece-commitment"
	app.Usage = ""
	app.Flags = flags
	app.Before = func(c *cli.Context) error {
		config.InitConfig(c.String("config"))
		handler.InitLog()
		inited.InitLog()
		return nil
	}
	app.Action = func(c *cli.Context) error {
		inited.InitApp()
		if err := handler.StartGenPieceCommitmentTask(); err != nil {
			panic(err)
		}
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(fmt.Sprintf("app run failed: %v\n", err.Error()))
	}
}
