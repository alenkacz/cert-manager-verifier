.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build -race -o ./bin/cm-verifier cmd/ctl/main.go