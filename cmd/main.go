package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"example/config"
	"example/server"
)

func main() {
	fs := flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ContinueOnError)
	opts, err := config.Configure(fs, os.Args[1:])
	if err != nil {
		config.PrintErrorAndDie(err)
	}

	s := server.NewServer().
		WorkerCount(opts.WorkerCount).
		Brokers(opts.Brokers).
		InTopic(opts.InTopic).
		OutTopic(opts.OutTopic).
		Banner(!opts.NoBanner).
		Verbose(opts.Verbose)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	wg.Add(1)
	go s.Run(ctx, cancel, &wg)

	<-ctx.Done()
	cancel()
	wg.Wait()
	log.Println("server is gracefully shutdown")
}
