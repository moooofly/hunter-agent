all: run

run: build
	./agent -H tcp://0.0.0.0:12345 -H unix:///var/run/hunter-agent.sock

build:
	go build cmd/agent/*.go

clean:
	rm agent
