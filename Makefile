test:
	go test -race ./...
.PHONY: test

lint:
	@if [ -z `which golangci-lint` ]; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b `go env GOPATH`/bin v1.50.1; \
	fi
	golangci-lint run ./...
.PHONY: lint
