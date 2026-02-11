include .local.env
export

export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING=${PG_URL}

.PHONY: run
run:
	go run .

.PHONY: test
test:
	go test -v -cover -race ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: gameserver
gameserver:
	air --build.cmd "go build -o ./tmp/gameserver ./cmd/gameserver" --build.bin "./tmp/gameserver"

.PHONY: client
client:
	go run ./cmd/client

.PHONY: clients
clients:
	go run ./cmd/client & go run ./cmd/client

.PHONY: wasm
wasm:
	mkdir -p ./web
	GOOS=js GOARCH=wasm go build -o ./web/client.wasm ./cmd/client
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" ./web/

.PHONY: web
web:
	air --build.cmd "make wasm && go build -o ./tmp/web ./cmd/web" --build.bin "./tmp/web"

.PHONY: goose-new
goose-new:
	@read -p "Enter the name of the new migration: " name; \
	goose -dir migrations create "$${name// /_}" sql

.PHONY: goose-up
goose-up:
	@echo "Running all new database migrations..."
	goose -dir migrations validate
	goose -dir migrations up

.PHONY: goose-down
goose-down:
	@echo "Running all down database migrations..."
	goose -dir migrations down

.PHONY: goose-reset
goose-reset:
	@echo "Dropping everything in database..."
	goose -dir migrations reset

.PHONY: goose-status
goose-status:
	goose -dir migrations status

