.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build -race -o ./bin/cm-verifier cmd/ctl/main.go

.PHONY: e2e-test
e2e-test:
	./hack/run-e2e-tests.sh