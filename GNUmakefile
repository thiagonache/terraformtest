TEST?=$$(go list ./...)

default: test

test: lint fmt
	go test -v -timeout=30s $(TEST)

actions:
	@echo "==> Running act with big image of ubuntu 18.04"
	@echo "==> Warning - this image is >18GB"
	act -P ubuntu-latest=nektos/act-environments-ubuntu:18.04

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -w -s .

lint:
	@echo "==> Checking source code against linters..."
	golint -set_exit_status ./...
