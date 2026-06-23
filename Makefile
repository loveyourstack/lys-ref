
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

# build MCP server
.PHONY: mcpb
mcpb:
	go build -o ./bin ./cmd/mcpsrv

# build ref server
.PHONY: srvb
srvb:
	go build -o ./bin ./cmd/refsrv

# build supplier server
.PHONY: suppsrvb
suppsrvb:
	go build -o ./bin ./cmd/suppsrv

# build all apps
.PHONY: build
build: clib mcpb srvb suppsrvb

###########################################################

# build and copy CLI to remote
.PHONY: copycli
copycli: clib
	rsync -az -e ssh ./bin/refcli lysref:/home/ubuntu/bin

# build and copy MCP server to remote
.PHONY: copymcp
copymcp: mcpb
	rsync -az -e ssh ./bin/mcpsrv lysref:/home/ubuntu/bin

# build and copy ref server to remote
.PHONY: copysrv
copysrv: srvb
	rsync -az -e ssh ./bin/refsrv lysref:/home/ubuntu/bin

# build and copy supplier server to remote
.PHONY: copysuppsrv
copysuppsrv: suppsrvb
	rsync -az -e ssh ./bin/suppsrv lysref:/home/ubuntu/bin

# build and copy ui/dist to remote
.PHONY: copyui
copyui:
	pnpm --dir frontend/lys-ref-ui build
	rsync --delete -az -e ssh ./frontend/lys-ref-ui/dist/ lysref:/home/ubuntu/frontend