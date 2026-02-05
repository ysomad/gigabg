.PHONY: run
run:
	go run .

.PHONY: test
test:
	go test -v -cover -race ./...

.PHONY: testcov
testcov:
	go test -v -race -covermode=atomic -coverprofile=coverage.out ./... && \
	go tool cover -html=coverage.out -o coverage.html

.PHONY: opencov
opencov:
	go tool cover -html=coverage.out

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative proto/game.proto

.PHONY: server
server:
	air --build.cmd "go build -o ./tmp/server ./cmd/server" --build.bin "./tmp/server"

.PHONY: client
client:
	go run ./cmd/client -server=localhost:8080
