package main

import (
	"os"
	"runtime"
	"time"

	log "tasadar.net/tionis/pipes/src/logger"
	"tasadar.net/tionis/pipes/src/server"

	"github.com/urfave/cli"
)

func main() {
	err := Run()
	if err != nil {
		log.Debug(err)
	}
}

// Version specifies the version
var Version string

// Run will run the command line program
func Run() (err error) {
	// use all the processors
	runtime.GOMAXPROCS(runtime.NumCPU())

	app := cli.NewApp()
	app.Name = "duct"
	if Version == "" {
		Version = "v1.0.0"
	}
	app.Version = Version
	app.Compiled = time.Now()
	app.Usage = "duct provides simple endpoints for managing data between applications"
	app.UsageText = ``
	app.Commands = []cli.Command{
		{
			Name:        "serve",
			Description: "start server for relaying data",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "port", Value: "9002", Usage: "port to use"},
			},
			HelpName: "duct serve",
			Action: func(c *cli.Context) error {
				setDebug(c)
				return server.Serve(c.String("port"))
			},
		},
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "debug", Usage: "toggle debug mode"},
	}
	app.HideHelp = false
	app.HideVersion = false

	return app.Run(os.Args)
}

func setDebug(c *cli.Context) {
	if c.GlobalBool("debug") {
		log.SetLevel("debug")
	} else {
		log.SetLevel("info")
	}
}
