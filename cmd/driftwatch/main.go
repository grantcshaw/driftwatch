package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/yourorg/driftwatch/internal/config"
)

const defaultConfigPath = "driftwatch.yaml"

func main() {
	configPath := flag.String("config", defaultConfigPath, "path to configuration file")
	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetPrefix("driftwatch ")

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	log.Printf("loaded %d environment(s), check interval %v",
		len(cfg.Environments), cfg.CheckInterval)

	for _, env := range cfg.Environments {
		log.Printf("  environment: name=%s provider=%s region=%s",
			env.Name, env.Provider, env.Region)
	}

	if cfg.Alerts.LogOnly {
		log.Println("alert mode: log-only (no external notifications)")
	} else if cfg.Alerts.SlackWebhook != "" {
		log.Println("alert mode: slack")
	} else {
		log.Println("alert mode: log-only (no webhook configured)")
	}

	// TODO: start drift detection loop
	log.Println("driftwatch started — drift detection not yet implemented")
	select {}
}
