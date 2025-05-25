MODULE := github.com/sejo412/ya-metrics/internal/config
BUILD_VERSION ?= 0.0.0-rc1
BUILD_COMMIT ?= $$(git rev-parse HEAD)
BUILD_DATE ?= $$(date -R)

.PHONY: all
all: proto server agent

.PHONY: server
server:
	go build -race -ldflags \
		"-X '$(MODULE).BuildVersion=$(BUILD_VERSION)'\
		-X '$(MODULE).BuildCommit=$(BUILD_COMMIT)'\
		-X '$(MODULE).BuildDate=$(BUILD_DATE)'"\
		-o ./cmd/server/server ./cmd/server/

.PHONY: agent
agent:
	go build -race -ldflags \
		"-X '$(MODULE).BuildVersion=$(BUILD_VERSION)'\
		-X '$(MODULE).BuildCommit=$(BUILD_COMMIT)'\
		-X '$(MODULE).BuildDate=$(BUILD_DATE)'"\
		-o ./cmd/agent/agent ./cmd/agent/

.PHONY: staticlint
staticlint:
	go build -o ./cmd/staticlint/staticlint ./cmd/staticlint/

.PHONY: tests
tests:
	echo "example: make tests t=TestIteration2"
	metricstest -test.v -test.run=^$(t)$$ -source-path=. -binary-path=./cmd/server/server -agent-binary-path=./cmd/agent/agent

.PHONY: statictest
statictest:
	go vet -vettool=$$(which statictest) ./...

.PHONY: statictest2
statictest2:
	go vet -vettool=./cmd/staticlint/staticlint ./...

.PHONY: cover
cover:
	#go test ./... -coverprofile=./coverage.out -covermode=atomic -coverpkg=./...
	go test -v -tags integration -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out
	@rm -f coverage.out

.PHONY: lint
lint:
	task lint

.PHONY: fieldalignment-diff
fieldalignment-diff:
	fieldalignment -fix -diff ./...

.PHONY: fieldalignment-fix
fieldalignment-fix:
	fieldalignment -fix ./...

.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto
