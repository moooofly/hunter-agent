package main

import (
	"io"
	"time"

	"github.com/Shopify/sarama"
	"github.com/census-instrumentation/opencensus-proto/gen-go/exporterproto"
	"github.com/census-instrumentation/opencensus-proto/gen-go/traceproto"
	"github.com/golang/protobuf/proto"
	"github.com/moooofly/hunter-agent/gen-go/dumpproto"
	"github.com/sirupsen/logrus"
)

type server struct {
	topic     string
	partition string
	producer  sarama.AsyncProducer

	done     chan struct{}
	pipeline chan *sarama.ProducerMessage
}

func newAsyncProducer(brokerList []string) sarama.AsyncProducer {

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
	config.Producer.Compression = sarama.CompressionSnappy   // Compress messages
	config.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms

	producer, err := sarama.NewAsyncProducer(brokerList, config)
	if err != nil {
		logrus.Fatalln("Failed to start Sarama producer:", err)
	}

	// We will just log to STDOUT if we're not able to produce messages.
	// Note: messages will only be returned here after all retry attempts are exhausted.
	go func() {
		for err := range producer.Errors() {
			logrus.Println("Failed to write access log entry:", err)
		}
	}()

	return producer
}

func newFlowControlServer(cli *DaemonCli) *server {
	p := newAsyncProducer(cli.Config.Brokers)

	s := &server{
		topic:     cli.Topic,
		partition: cli.Partition,
		producer:  p,
		done:      make(chan struct{}),
		pipeline:  make(chan *sarama.ProducerMessage, cli.QueueSize),
	}

	s.launch()

	return s
}

func (s *server) ExportMetrics(stream exporterproto.Export_ExportMetricsServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		logrus.Debug(in)
	}
}

func (s *server) ExportSpan(stream exporterproto.Export_ExportSpanServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// FIXME: add debug switch
		/*
			for _, sp := range in.Spans {
				logrus.Debugf("---> span: %v\n", sp)
			}
		*/

		var key sarama.Encoder
		if s.partition != "" {
			key = sarama.StringEncoder(s.partition)
			// FIXME: add debug switch
			//logrus.Debugf("partition: %s", s.partition)
		} else {
			key = sarama.ByteEncoder(in.Spans[0].GetTraceId())
			// FIXME: add debug switch
			//logrus.Debugf("partition: %s", fmt.Sprintf("%02x", in.Spans[0].GetTraceId()[:]))
		}

		// FIXME: add debug switch
		//logrus.Debugf("len(in.Spans): %d\n", len(in.Spans))

		dump, err := dumpSpans(in.Spans)
		if err != nil {
			logrus.Errorf("dumpSpans err: %v", err)
			return err
		}

		s.push(&sarama.ProducerMessage{
			Topic: s.topic,
			Key:   key,
			Value: sarama.ByteEncoder(dump),
		})
	}
}

func dumpSpans(spans []*traceproto.Span) ([]byte, error) {
	ds := &dumpproto.DumpSpans{Spans: spans}
	serialized, err := proto.Marshal(ds)
	if err != nil {
		return nil, err
	}
	return serialized, nil
}

func (s *server) push(ms *sarama.ProducerMessage) bool {
	select {
	case s.pipeline <- ms:
		msgIn.Add(1)
		return true
	default:
		// drop message if queue is full
		msgDrop.Add(1)
		return false
	}
}

func (s *server) launch() {
	go func() {
		for {
			select {
			case <-s.done:
				return
			case ms := <-s.pipeline:
				msgOut.Add(1)
				s.onMessage(ms)
			}
		}
	}()
}

func (s *server) stop() {
	close(s.done)
}

func (s *server) onMessage(ms *sarama.ProducerMessage) {
	s.producer.Input() <- ms
}
