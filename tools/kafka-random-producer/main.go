// This package is used for only local tests.
// Do not use at the Prod!
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	kafka "github.com/teamseodo/kafka-do"
)

type Request struct {
	Id int `json:"id"`
}

var (
	batchCount = flag.Int("batch-count", 100, "")
	kafkaUrl   = flag.String("kafka-url", "kafka:9092", "")
	waitTime   = flag.Int("wait-time-seconds", 5*60, "")
	verbose    = flag.Bool("verbose", false, "")
)

// Produces random keywords as much as you want.
func main() {
	flag.Parse()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	go run(ctx, cancel)
	<-ctx.Done()

	log.Println("server is gracefully shutdown")
}

func run(ctx context.Context, cancel context.CancelFunc) {
	kkw, err := kafka.NewProducer(strings.Split(*kafkaUrl, ","), 5)
	if err != nil {
		log.Fatal(err)
	}
	defer kkw.Close()

	for {
		rand.Seed(time.Now().UnixNano())
		var messages []*sarama.ProducerMessage
		for i := 0; i < *batchCount; i++ {
			req := Request{
				Id: rand.Intn(1000) + 1000,
			}
			if *verbose {
				log.Println(req)
			}

			reqJson, err := json.Marshal(&req)
			if err != nil {
				log.Print(err)
				continue
			}

			messages = append(messages, &sarama.ProducerMessage{
				Value: sarama.StringEncoder(reqJson),
				Topic: "requests",
			})
		}

		err = kafka.ProduceBatch(ctx, kkw, messages, "")
		if err != nil {
			log.Print(err.Error())
		}

		if *waitTime == 0 {
			cancel()
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(*waitTime) * time.Second):
		}
	}
}
