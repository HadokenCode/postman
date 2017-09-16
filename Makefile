BINARY=postman

VERSION=0.1
BUILD=`git rev-parse HEAD | head -c 8`

LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"

.PHONY: build
build:
	@echo "==> Building"
	@go build ${LDFLAGS} -o build/${BINARY} cmd/*.go
	@echo "\n==>\033[32m Ok\033[m\n"

.PHONY: install
install:
	@echo "==> Installing in ${GOPATH}/bin/${BINARY}"
	@cp build/${BINARY} ${GOPATH}/bin/
	@echo "\n==>\033[32m Ok\033[m\n"

.PHONY: test
test:
	@echo "==> Running all tests"
	@go test ./...

.PHONY: lint
lint:
	@echo "==> Running static analysis tests"
	@golint -set_exit_status -min_confidence 0.9 cmd/...
	@golint -set_exit_status -min_confidence 0.9 async/...
	@golint -set_exit_status -min_confidence 0.9 proxy/...

.PHONY: test.setup
test.setup:
	@echo "==> Install dep"
	@go get github.com/golang/dep/cmd/dep
	@echo "==> Install golint"
	@go get github.com/golang/lint
	@echo "==> Install dependencies"
	@dep ensure

# Show to-do items per file.
todo:
	@grep \
		--exclude-dir=vendor \
		--exclude-dir=node_modules \
		--text \
		--color \
		-nRo -E ' TODO:.*|SkipNow' .
	