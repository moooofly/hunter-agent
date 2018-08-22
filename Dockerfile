FROM golang:1.9.0 as builder
WORKDIR /go/src/github.com/moooofly/hunter-agent
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o agent cmd/agent/*.go

# Final image.
FROM scratch
LABEL maintainer "moooofly <centos.sf@gmail.com>"
COPY --from=builder /go/src/github.com/moooofly/hunter-agent/agent .
COPY agent.json.template /etc/hunter/agent.json
# Usage: docker run -it -p 12345:12345 -v /var/run:/var/run --rm hunter-agent:1726446 -H tcp://0.0.0.0:12345 -H unix:///var/run/hunter-agent.sock --metrics-addr 0.0.0.0:12346 --broker 10.1.8.95:9092 --topic jaeger-spans-test-001
ENTRYPOINT ["/agent"]
CMD ["-h"]
