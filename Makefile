.PHONY: all
all: server agent

.PHONY: server
server:
	go build -o ./cmd/server/server ./cmd/server/main.go

.PHONY: agent
agent:
	go build -o ./cmd/agent/agent ./cmd/agent/main.go

.PHONY: tests
tests:
	metricstest -test.v -test.run=^$(t)$$ -binary-path=./cmd/server/server -agent-binary-path=./cmd/agent/agent
