package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"example/config"

	"github.com/Shopify/sarama"
	"github.com/getsentry/sentry-go"
	kafka "github.com/teamseodo/kafka-do"
)

type Server interface {
	WorkerCount(int) Server
	Brokers([]string) Server
	InTopic(string) Server
	OutTopic(string) Server
	Banner(bool) Server
	Verbose(bool) Server
	Run(context.Context, context.CancelFunc, *sync.WaitGroup)
}

type server struct {
	workerCount int
	brokers     []string
	inTopic     string
	outTopic    string
	banner      bool
	verbose     bool
}

// NewServer returns a new server.
func NewServer() Server {
	s := server{}
	return &s
}

// Brokers sets kafka brokers.
func (s *server) Brokers(k []string) Server {
	s.brokers = k
	return s
}

// InTopic sets topic name for input messages.
func (s *server) InTopic(k string) Server {
	s.inTopic = k
	return s
}

// OutTopic sets topic name for output messages.
func (s *server) OutTopic(k string) Server {
	s.outTopic = k
	return s
}

// WorkerCount sets how many worker will work.
func (s *server) WorkerCount(count int) Server {
	s.workerCount = count
	return s
}

// Verbose sets verbose feature.
func (s *server) Verbose(verbose bool) Server {
	s.verbose = verbose
	return s
}

// Banner sets banner on/off.
func (s *server) Banner(banner bool) Server {
	s.banner = banner
	return s
}

// Run starts server.
func (s *server) Run(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	if s.banner {
		fmt.Println(config.Banner)
	}

	producer, err := kafka.NewProducer(s.brokers, 5)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	sentryOk := false
	if key := os.Getenv("SENTRY_KEY"); key != "" {
		err := sentry.Init(sentry.ClientOptions{Dsn: key})
		if err != nil {
			log.Printf("sentry init err: %s", err)
		} else {
			log.Print("sentry is set")
			sentryOk = true
		}
	}

	// recover panic errors and log them.
	// if sentry is ok, sent them to sentry too.
	defer func() {
		if err := recover(); err != nil {
			p := fmt.Sprintf("panic: %s\nstack trace: %s", err, string(debug.Stack()))
			log.Print(p)
			if sentryOk {
				sentry.CurrentHub().Recover(p)
				sentry.Flush(time.Second * 5)
			}
		}
	}()

	var consumer sarama.ConsumerGroup
	for {
		consumer, err = kafka.NewConsumerGroup(s.brokers, "cosmos")
		if err != nil {
			log.Printf("consumer group creation err: %s", err)
			select {
			case <-ctx.Done():
				wg.Done()
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}
		defer consumer.Close()
		break
	}

	time.Sleep(2 * time.Second)
	log.Println("server is started")

	consumeChan := make(chan sarama.ConsumerMessage, s.workerCount)
	defer close(consumeChan)
	produceChan := make(chan interface{}, s.workerCount)
	defer close(consumeChan)
	produceErrChan := make(chan error, s.workerCount)
	defer close(produceErrChan)

	wg.Add(2)
	go kafka.ConsumeChan(ctx, wg, consumer, []string{s.inTopic}, consumeChan)
	go kafka.ProduceChan(ctx, wg, producer, produceChan, produceErrChan, "responses")
	go logErrors(ctx, produceErrChan)

	// HERE DO YOUR TASK.
	for i := 0; i < s.workerCount; i++ {
		wg.Add(1)
		go print(ctx, wg, consumeChan, produceChan)
	}

	<-ctx.Done()
	if len(consumeChan) > 0 { // produce back if there is any message.
		err = kafka.ProduceBatch(ctx, producer, consumeChanToSlice(consumeChan), s.outTopic)
		if err != nil {
			err := fmt.Sprintf("produce to in topic error: %s", err)
			log.Print(err)
			if sentryOk {
				sentry.CaptureMessage(err)
			}
		}
	}
	wg.Done()
}

// logErrors reads from errs channel and write them to stdout as a log.
func logErrors(ctx context.Context, errs chan error) {
	for {
		select {
		case err := <-errs:
			log.Printf("error: %s", err)
		case <-ctx.Done():
			return
		}
	}
}

// consumeChanToSlice converts chan sarama.ConsumerMessage to a slice.
func consumeChanToSlice(ch chan sarama.ConsumerMessage) (chs []sarama.ConsumerMessage) {
	for {
		v, ok := <-ch
		if !ok {
			return chs
		}
		chs = append(chs, v)
	}
}

// print read from cn channel, writes to strdout and write to pr channel.
func print(ctx context.Context, wg *sync.WaitGroup, cn chan sarama.ConsumerMessage, pr chan interface{}) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-cn:
			fmt.Printf("message: %s %s\n", m.Timestamp, m.Value)
			pr <- m
		}
	}
}
