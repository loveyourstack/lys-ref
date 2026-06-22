package myapp

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/loveyourstack/lys/lyspgdb"
)

// McpConfig contains all configuration settings for the MCP server.
type McpConfig struct {
	Db        lyspgdb.Database `toml:"database"`
	DbMcpUser lyspgdb.User
}

func (c *McpConfig) LoadFromFile(configFilePath string) (err error) {

	// ensure supplied path exists
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return fmt.Errorf("configFilePath does not exist: %s", configFilePath)
	} else if err != nil {
		return fmt.Errorf("os.Stat failed: %w", err)
	}

	// read conf from toml file
	if _, err := toml.DecodeFile(configFilePath, c); err != nil {
		return fmt.Errorf("toml.DecodeFile failed: %w", err)
	}

	// validate conf
	if c.Db.Database == "" {
		return fmt.Errorf("MCP config: Db.Database is empty")
	}
	if c.DbMcpUser.Name == "" {
		return fmt.Errorf("MCP config: DbMcpUser.Name is empty")
	}

	return nil
}
