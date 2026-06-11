package myapp

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/loveyourstack/connectors/maxmind/mmapi"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys-ref/pkg/aws/awsapi"
	"github.com/loveyourstack/lys/lysmail"
	"github.com/loveyourstack/lys/lyspgdb"
	"golang.org/x/crypto/bcrypt"
)

// general contains the general application config
type general struct {
	AppName       string
	Env           appenv.Enum
	Debug         bool
	DownloadsPath string
}

// api contains the API config
type api struct {
	Port              string
	UseAuthentication bool
}

// ui contains the config for the UI which accesses the API
type ui struct {
	Url string
}

// developer contains details of the developer. Used to create the initial system.user on db creation
type developer struct {
	GivenName  string
	FamilyName string
	Email      string
	Password   string
}

// Process contains the config for the process schema
type Process struct {
	CliCmdPrefix string
}

// GetReplacements sets the string replacements from the given developer config to be used on db creation
func (d developer) GetReplacements() (fileReplacements []lyspgdb.FileReplacement, err error) {

	hashedPw, err := bcrypt.GenerateFromPassword([]byte(d.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("bcrypt.GenerateFromPassword failed: %w", err)
	}

	fileReplacements = []lyspgdb.FileReplacement{
		{
			From: "<developer_givenName>",
			To:   d.GivenName,
		},
		{
			From: "<developer_familyName>",
			To:   d.FamilyName,
		},
		{
			From: "<developer_email>",
			To:   d.Email,
		},
		{
			From: "<developer_hashedPw>",
			To:   string(hashedPw),
		},
	}

	return fileReplacements, nil
}

// Config contains all configuration settings
type Config struct {
	General      general
	Db           lyspgdb.Database `toml:"database"`
	DbSuperUser  lyspgdb.User
	DbOwnerUser  lyspgdb.User
	DbServerUser lyspgdb.User
	DbCliUser    lyspgdb.User
	API          api
	UI           ui
	Developer    developer

	Aws     awsapi.Conf
	MaxMind mmapi.Conf
	Process Process
	Smtp    lysmail.SmtpConfig
}

func (c *Config) LoadFromFile(configFilePath string) (err error) {

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

	// enforce dev-only rules
	if c.General.Env != appenv.Dev {
		if !c.API.UseAuthentication {
			return fmt.Errorf("config validation failed: API.UseAuthentication cannot be false when General.Env is not dev")
		}
	}

	err = c.Smtp.Validate()
	if err != nil {
		return fmt.Errorf("c.Smtp.Validate failed: %w", err)
	}

	return nil
}
