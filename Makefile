.PHONY: all
all: server agent

.PHONY: server
server:
	go build -o ./cmd/server/server ./cmd/server/

.PHONY: agent
agent:
	go build -o ./cmd/agent/agent ./cmd/agent/

.PHONY: tests
tests:
	echo "example: make tests t=TestIteration2"
	metricstest -test.v -test.run=^$(t)$$ -source-path=. -binary-path=./cmd/server/server -agent-binary-path=./cmd/agent/agent

.PHONY: statictest
statictest:
	go vet -vettool=$(which statictest) ./...

.PHONY: lint
lint:
	task lint
