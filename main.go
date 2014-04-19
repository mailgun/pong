package main

import (
	"fmt"
	"github.com/mailgun/cli"
	log "github.com/mailgun/gotools-log"
	"github.com/mailgun/vulcanb/config"
	"github.com/mailgun/vulcanb/model"
	"os"
	"os/signal"
)

func main() {
	app := cli.NewApp()
	app.Name = "vulcanb"
	app.Usage = "Command line tool that generates endpoints with different behavior for testing purposes"
	app.Flags = []cli.Flag{
		cli.StringFlag{"c, config", "", "Yaml file with endpoint specifications"},
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
