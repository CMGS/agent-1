package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/projecteru2/agent/api"
	"github.com/projecteru2/agent/common"
	"github.com/projecteru2/agent/engine"
	"github.com/projecteru2/agent/types"
	"github.com/projecteru2/agent/utils"
	"github.com/projecteru2/agent/watcher"
	"gopkg.in/urfave/cli.v1"
)

func setupLogLevel(l string) error {
	level, err := log.ParseLevel(l)
	if err != nil {
		return err
	}
	log.SetLevel(level)
	return nil
}

func initConfig(c *cli.Context) *types.Config {
	config := &types.Config{}
	err := config.LoadConfigFromFile(c.String("config"))
	if err != nil {
		log.Warnf("[main] load config failed %v", err)
	}
	config.PrepareConfig(c)
	return config
}

func serve(c *cli.Context) error {
	if err := setupLogLevel(c.String("log-level")); err != nil {
		log.Fatal(err)
	}

	config := initConfig(c)
	log.Debugf("[config] %v", config)
	utils.WritePid(config.PidFile)
	defer os.Remove(config.PidFile)

	watcher.InitMonitor()
	go watcher.LogMonitor.Serve()

	agent, err := engine.NewEngine(config)
	if err != nil {
		log.Fatal(err)
	}

	go api.Serve(config.API.Addr)

	if err := agent.Run(); err != nil {
		log.Fatalf("Agent caught error %s", err)
	}
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "Eru-Agent"
	app.Usage = "Run eru agent"
	app.Version = common.ERU_AGENT_VERSION
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config",
			Value:  "/etc/eru/agent.yaml",
			Usage:  "config file path for agent, in yaml",
			EnvVar: "ERU_AGENT_CONFIG_PATH",
		},
		cli.StringFlag{
			Name:   "log-level",
			Value:  "INFO",
			Usage:  "set log level",
			EnvVar: "ERU_AGENT_LOG_LEVEL",
		},
		cli.StringFlag{
			Name:   "core-endpoint",
			Value:  "",
			Usage:  "core endpoint",
			EnvVar: "ERU_AGENT_CORE_ENDPOINT",
		},
		cli.StringFlag{
			Name:   "docker-endpoint",
			Value:  "",
			Usage:  "docker endpoint",
			EnvVar: "ERU_AGENT_DOCKER_ENDPOINT",
		},
		cli.Int64Flag{
			Name:   "metrics-step",
			Value:  0,
			Usage:  "interval for metrics to send",
			EnvVar: "ERU_AGENT_METRICS_STEP",
		},
		cli.StringSliceFlag{
			Name:   "metrics-transfers",
			Value:  &cli.StringSlice{},
			Usage:  "metrics destinations",
			EnvVar: "ERU_AGENT_METRICS_TRANSFERS",
		},
		cli.StringFlag{
			Name:   "api-addr",
			Value:  "",
			Usage:  "agent API serving address",
			EnvVar: "ERU_AGENT_API_ADDR",
		},
		cli.StringSliceFlag{
			Name:   "log-forwards",
			Value:  &cli.StringSlice{},
			Usage:  "log destinations",
			EnvVar: "ERU_AGENT_LOG_FORWARDS",
		},
		cli.StringFlag{
			Name:   "log-stdout",
			Value:  "",
			Usage:  "forward stdout out? yes/no",
			EnvVar: "ERU_AGENT_LOG_STDOUT",
		},
		cli.StringFlag{
			Name:   "pidfile",
			Value:  "",
			Usage:  "pidfile to save",
			EnvVar: "ERU_AGENT_PIDFILE",
		},
		cli.IntFlag{
			Name:   "health-check-interval",
			Value:  0,
			Usage:  "interval for agent to check container's health status",
			EnvVar: "ERU_AGENT_HEALTH_CHECK_INTERVAL",
		},
		cli.IntFlag{
			Name:   "health-check-timeout",
			Value:  0,
			Usage:  "timeout for agent to check container's health status",
			EnvVar: "ERU_AGENT_HEALTH_CHECK_TIMEOUT",
		},
	}
	app.Action = serve
	app.Run(os.Args)
}
