// This package is used for only local tests.
// Do not use at the Prod!
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	kafka "github.com/teamseodo/kafka-do"
)

type Response struct {
	Id int `json:"id"`
}

var (
	batchCount = flag.Int("batch-count", 100, "")
	kafkaUrl   = flag.String("kafka-url", "kafka:9092", "")
)

// Reads serps from Serps Kafka and write them to results/ folder.
func main() {
	flag.Parse()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	go run(ctx)
	<-ctx.Done()
	cancel()
	log.Println("server is gracefully shutdown")

}

func run(ctx context.Context) {
	if !isExist("./results") {
		err := createFolder("./results")
		if err != nil {
			log.Fatalf("can't create results folder, error: %s", err)
		}
	}

	ksr, err := kafka.NewConsumerGroup(strings.Split(*kafkaUrl, ","), "device")
	if err != nil {
		log.Fatal(err)
	}
	defer ksr.Close()

	for {
		if ctx.Err() != nil {
			return
		}

		messages := kafka.ConsumeBatch(ctx, ksr, []string{"responses"}, *batchCount)

		var wg sync.WaitGroup
		for _, m := range messages {
			wg.Add(1)
			go writeToDevice(ctx, &wg, m.Value)
		}
		wg.Wait()

		log.Printf("wrote %d results to device", len(messages))
	}
}

func writeToDevice(ctx context.Context, wg *sync.WaitGroup, file []byte) {
	defer wg.Done()

	var r Response
	err := json.Unmarshal(file, &r)
	if err != nil {
		log.Println("JSON Decode Error:", err.Error())
		return
	}

	fj, err := os.Create(fmt.Sprintf("results/%d.json", r.Id))
	if err != nil {
		log.Printf("json file write error: %s", err.Error())
		return
	}

	fj.Write(file)
}

func isExist(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

func createFolder(folder string) error {
	return os.Mkdir(folder, 0744)
}
