package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys/lyspgdb"
)

func main() {
	configFilePath := flag.String("config", "mcp_config.toml", "Path to the config file")
	flag.Parse()

	if err := run(*configFilePath); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(configFilePath string) error {

	// load config
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", configFilePath)
	}
	var conf myapp.McpConfig
	if _, err := toml.DecodeFile(configFilePath, &conf); err != nil {
		return fmt.Errorf("toml.DecodeFile failed: %w", err)
	}

	// open DB connection pool
	db, err := lyspgdb.GetPool(context.Background(), conf.Db, conf.DbMcpUser, "lys-ref MCP")
	if err != nil {
		return fmt.Errorf("lyspgdb.GetPool failed: %w", err)
	}
	defer db.Close()

	if err := runServer(db); err != nil {
		return fmt.Errorf("mcp server error: %w", err)
	}

	return nil
}
