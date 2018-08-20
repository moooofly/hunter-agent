BROKERS := $(shell docker port kafka-docker_kafka_1 9092/tcp)

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
	go build cmd/agent/*.go

clean:
	rm -f agent