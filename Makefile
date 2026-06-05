
# variables
TODAY         := $(shell date +%Y-%m-%d)
ONE_MONTH_AGO := $(shell date -d '1 month ago' +%Y-%m-%d)

# install CLI (must use this when testing so that latest config is always used)
.PHONY: cli
cli:
	sudo cp ref_config.toml /usr/local/etc
	go install ./cmd/refcli

# (re-)create database
.PHONY: db
db: cli
	refcli createDb
	refcli dm aggCampPerf
	refcli admin installLysPgMon
	refcli admin checkDb
	refcli pub addFakeUpdates
	refcli ecb sync currencies
	refcli ecb sync xr $(ONE_MONTH_AGO) $(TODAY)

# reset db data
.PHONY: resetdbdata
resetdbdata: cli
	refcli resetDbData
	refcli dm aggCampPerf
	refcli admin checkDb

###########################################################

# ref server start
.PHONY: srv
srv:
	go run ./cmd/refsrv

# Supplier API server start
.PHONY: suppsrv
suppsrv:
	go run ./cmd/suppsrv

###########################################################

# run all tests
.PHONY: tests
tests: 
	go test -race ./...

###########################################################

# build CLI
.PHONY: clib
clib:
	go build -o ./bin ./cmd/refcli


# build server
.PHONY: srvb
srvb:
	go build -o ./bin ./cmd/refsrv

# build supplier server
.PHONY: suppsrvb
suppsrvb:
	go build -o ./bin ./cmd/suppsrv

# build all apps
.PHONY: build
build: clib srvb suppsrvb