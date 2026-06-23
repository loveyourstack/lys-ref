package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys/lyspgdb"
)

func main() {
	configFilePath := flag.String("config", "mcp_config.toml", "Path to the config file")
	flag.Parse()

	// load config
	if _, err := os.Stat(*configFilePath); os.IsNotExist(err) {
		log.Fatalf("mcp initialization: config file not found: %s", *configFilePath)
	}
	var cfg myapp.McpConfig
	if _, err := toml.DecodeFile(*configFilePath, &cfg); err != nil {
		log.Fatalf("mcp initialization: failed to decode config: %s", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// open DB connection pool
	db, err := lyspgdb.GetPool(ctx, cfg.Db, cfg.DbMcpUser, "lys-ref MCP")
	if err != nil {
		log.Fatalf("mcp initialization: lyspgdb.GetPool failed: %s", err)
	}
	defer db.Close()

	if err := runServer(db); err != nil {
		log.Fatalf("mcp server error: %s", err)
	}
}
