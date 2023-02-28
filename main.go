package main

import (
	"fmt"
	"gen-piece-commitment/config"
	"gen-piece-commitment/handler"
	"gen-piece-commitment/inited"
	"github.com/urfave/cli/v2"
	"os"
)

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "config",
			Aliases:  []string{"c", "conf"},
			Value:    "config.toml",
			Required: true,
			Usage:    "--config=config.toml",
		},
	},
	Before: func(c *cli.Context) error {
		config.InitConfig(c.String("config"))
		handler.InitLog()
		inited.InitLog()
		inited.InitApp()
		return nil
	},
	Action: func(c *cli.Context) error {
		if err := handler.StartGenPieceCommitmentTask(); err != nil {
			panic(err)
		}
		return nil
	},
}

var importCmd = &cli.Command{
	Name:  "import",
	Usage: "",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "config",
			Aliases:  []string{"c", "conf"},
			Value:    "config.toml",
			Required: true,
			Usage:    "--config=config.toml",
		},
		&cli.StringFlag{
			Name:     "miner",
			Required: false,
			Usage:    "--miner=t01000",
		},
		&cli.StringFlag{
			Name:     "file",
			Required: true,
			Usage:    "--file=import.txt (miner proposalCid carPath)",
		},
	},
	Before: func(c *cli.Context) error {
		config.InitConfig(c.String("config"))
		handler.InitLog()
		inited.InitLog()
		inited.InitApp()
		return nil
	},
	Action: func(c *cli.Context) error {
		if err := handler.ImportDeal(c.String("miner"), c.String("file")); err != nil {
			panic(err)
		}
		return nil
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "gen-piece-commitment"
	app.Usage = ""
	app.Commands = []*cli.Command{
		runCmd,
		importCmd,
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(fmt.Sprintf("app run failed: %v\n", err.Error()))
	}
}
