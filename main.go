package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/mailgun/log"
	"github.com/mailgun/pong/config"
	"github.com/mailgun/pong/model"
	"os"
	"os/signal"
)

func main() {
	app := cli.NewApp()
	app.Name = "pong"
	app.Usage = "Command line tool that generates endpoints with different behavior for testing purposes"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "c, config", Usage: "Yaml file with endpoint specifications"},
	}
	app.Action = startService
	app.Run(os.Args)
}

func startService(c *cli.Context) {
	servers, logConfig, err := config.ParseConfig(c.String("config"))
	if err != nil {
		fmt.Println("Failed to load config file '%s' err:", c.String("config"), err)
		return
	}
	log.Init(logConfig)
	for _, s := range servers {
		go model.StartServer(s)
	}
	inChan := make(chan os.Signal, 1)
	signal.Notify(inChan, os.Interrupt, os.Kill)
	log.Infof("Got signal %s, exiting now", <-inChan)
}
