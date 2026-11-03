package main

import (
	"testing"

	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys-ref/internal/myapp"
)

func TestNewBaseApplication(t *testing.T) {
	tests := []struct {
		name             string
		env              appenv.Enum
		wantRealIPHeader bool
	}{
		{name: "dev", env: appenv.Dev},
		{name: "prod", env: appenv.Prod, wantRealIPHeader: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := &myapp.Config{}
			conf.General.Env = tt.env
			conf.RpcClient.ApiKey = "test-api-key"

			srvApp, err := newBaseApplication(conf)
			if err != nil {
				t.Fatalf("newBaseApplication() error = %v", err)
			}

			if srvApp.Config != conf {
				t.Error("Config was not set to the provided config")
			}
			if srvApp.Logger == nil {
				t.Error("Logger was not initialized")
			}
			if srvApp.Validate == nil {
				t.Error("Validate was not initialized")
			}
			if srvApp.GetOptions.CsvDelimiter != '|' {
				t.Errorf("GetOptions.CsvDelimiter = %q, want '|'", srvApp.GetOptions.CsvDelimiter)
			}
			if srvApp.PostOptions.MaxBodySize <= 0 || srvApp.PostOptions.MaxImportRecs <= 0 {
				t.Errorf("PostOptions were not initialized with defaults: %+v", srvApp.PostOptions)
			}
			if srvApp.AuthedRateLimits == nil || srvApp.UnauthedRateLimits == nil {
				t.Error("rate limits were not initialized")
			}
			if srvApp.LoginAttempts == nil || srvApp.Sessions == nil {
				t.Error("authentication state was not initialized")
			}
			if srvApp.UseRealIpHeader != tt.wantRealIPHeader {
				t.Errorf("UseRealIpHeader = %t, want %t", srvApp.UseRealIpHeader, tt.wantRealIPHeader)
			}
			if srvApp.RpcApiKey != "test-api-key" {
				t.Errorf("RpcApiKey = %q, want %q", srvApp.RpcApiKey, "test-api-key")
			}
			if srvApp.ShutdownTimeouts != defaultShutdownTimeouts {
				t.Errorf("ShutdownTimeouts = %+v, want %+v", srvApp.ShutdownTimeouts, defaultShutdownTimeouts)
			}
			if srvApp.Db != nil || srvApp.OwnerDb != nil || srvApp.rpcClient != nil {
				t.Error("base application unexpectedly initialized a database or RPC connection")
			}

			srvApp.Close()
		})
	}
}
