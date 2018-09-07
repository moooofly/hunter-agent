BROKERS := $(shell docker port kafka-docker_kafka_1 9092/tcp)

LDFLAGS += -X "github.com/moooofly/hunter-agent/version.GitCommit=$(shell git rev-parse --short HEAD)"
LDFLAGS += -X "github.com/moooofly/hunter-agent/version.Version=$(shell cat VERSION)"
LDFLAGS += -X "github.com/moooofly/hunter-agent/version.BuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S')"

all:
	@echo "Usage:"
	@echo "  1. make local"
	@echo "  2. make dev"

local: build
	@# local env topic: test
	./agent -H tcp://0.0.0.0:12345 -H unix:///var/run/hunter-agent.sock --metrics-addr 0.0.0.0:12346 --broker ${BROKERS} --topic test

dev: build
	@# dev env topic: jaeger-spans-test-001
	./agent -H tcp://0.0.0.0:12345 -H unix:///var/run/hunter-agent.sock --metrics-addr 0.0.0.0:12346 --broker 10.1.8.95:9092 --topic jaeger-spans-test-001

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o agent cmd/agent/*.go
	@# not support MacOS yet
	@#CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o agent.mac cmd/agent/*.go

clean:
	rm -f agent

docker:
	docker build -t hunter-agent:$(shell git rev-parse --short HEAD) .

docker_run:
	docker run -it -p 12345:12345 -v /var/run:/var/run --rm hunter-agent:1726446 -H tcp://0.0.0.0:12345 -H unix:///var/run/hunter-agent.sock --metrics-addr 0.0.0.0:12346 --broker 10.1.8.95:9092 --topic jaeger-spans-test-001
