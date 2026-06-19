package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/loveyourstack/lys/lyspgdb"
)

// mcpConfig holds the minimal config required by the MCP server.
type mcpConfig struct {
	Db           lyspgdb.Database `toml:"database"`
	DbServerUser lyspgdb.User
}

func main() {
	configFilePath := flag.String("config", "ref_config.toml", "Path to the config file")
	flag.Parse()

	// load config
	if _, err := os.Stat(*configFilePath); os.IsNotExist(err) {
		log.Fatalf("config file not found: %s", *configFilePath)
	}
	var cfg mcpConfig
	if _, err := toml.DecodeFile(*configFilePath, &cfg); err != nil {
		log.Fatalf("failed to decode config: %s", err)
	}
	if cfg.Db.Database == "" {
		log.Fatalf("config: Db.Database is empty")
	}
	if cfg.DbServerUser.Name == "" {
		log.Fatalf("config: DbServerUser.Name is empty")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// open DB connection pool
	db, err := lyspgdb.GetPool(ctx, cfg.Db, cfg.DbServerUser, "lys-ref MCP")
	if err != nil {
		log.Fatalf("failed to open DB pool: %s", err)
	}
	defer db.Close()

	if err := runServer(db); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %s\n", err)
		os.Exit(1)
	}
}
