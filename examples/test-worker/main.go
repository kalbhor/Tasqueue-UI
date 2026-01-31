package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/kalbhor/tasqueue/v2"
	rb "github.com/kalbhor/tasqueue/v2/brokers/redis"
	rr "github.com/kalbhor/tasqueue/v2/results/redis"
)

type Payload struct {
	Arg1 int `json:"arg1"`
	Arg2 int `json:"arg2"`
}

type Result struct {
	Result int `json:"result"`
}

// AddProcessor adds two numbers
func AddProcessor(b []byte, ctx tasqueue.JobCtx) error {
	var pl Payload
	if err := json.Unmarshal(b, &pl); err != nil {
		return err
	}

	time.Sleep(time.Second * 2) // Simulate work

	rs, err := json.Marshal(Result{Result: pl.Arg1 + pl.Arg2})
	if err != nil {
		return err
	}

	ctx.Save(rs)
	log.Printf("✓ Added %d + %d = %d (Job ID: %s)", pl.Arg1, pl.Arg2, pl.Arg1+pl.Arg2, ctx.Meta.ID)
	return nil
}

// MultiplyProcessor multiplies two numbers
func MultiplyProcessor(b []byte, ctx tasqueue.JobCtx) error {
	var pl Payload
	if err := json.Unmarshal(b, &pl); err != nil {
		return err
	}

	time.Sleep(time.Second * 1) // Simulate work

	rs, err := json.Marshal(Result{Result: pl.Arg1 * pl.Arg2})
	if err != nil {
		return err
	}

	ctx.Save(rs)
	log.Printf("✓ Multiplied %d * %d = %d (Job ID: %s)", pl.Arg1, pl.Arg2, pl.Arg1*pl.Arg2, ctx.Meta.ID)
	return nil
}

// FailProcessor always fails (for testing failed jobs)
func FailProcessor(b []byte, ctx tasqueue.JobCtx) error {
	var pl Payload
	if err := json.Unmarshal(b, &pl); err != nil {
		return err
	}

	log.Printf("✗ Failing job on purpose (Job ID: %s)", ctx.Meta.ID)
	return errors.New("intentional failure for testing")
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	lo := slog.Default()

	// Get Redis address from environment variable or use default
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}

	log.Printf("Connecting to Redis at %s", redisAddr)

	// Create server with Redis broker and results backend
	srv, err := tasqueue.NewServer(tasqueue.ServerOpts{
		Broker: rb.New(rb.Options{
			Addrs:    []string{redisAddr},
			Password: "",
			DB:       0,
		}, lo),
		Results: rr.New(rr.Options{
			Addrs:    []string{redisAddr},
			Password: "",
			DB:       0,
		}, lo),
		Logger: lo.Handler(),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Register tasks
	log.Println("Registering tasks...")
	if err := srv.RegisterTask("add", AddProcessor, tasqueue.TaskOpts{Concurrency: 3}); err != nil {
		log.Fatal(err)
	}
	if err := srv.RegisterTask("multiply", MultiplyProcessor, tasqueue.TaskOpts{Concurrency: 3}); err != nil {
		log.Fatal(err)
	}
	if err := srv.RegisterTask("fail", FailProcessor, tasqueue.TaskOpts{Concurrency: 3}); err != nil {
		log.Fatal(err)
	}

	// Start the worker
	log.Println("Starting worker...")
	go srv.Start(ctx)

	// Wait a bit for worker to start
	time.Sleep(time.Second * 1)

	// Enqueue some individual jobs
	log.Println("\n=== Enqueueing individual jobs ===")

	// Successful jobs
	for i := 0; i < 5; i++ {
		b, _ := json.Marshal(Payload{Arg1: i, Arg2: i + 1})
		job, err := tasqueue.NewJob("add", b, tasqueue.JobOpts{})
		if err != nil {
			log.Fatal(err)
		}
		id, err := srv.Enqueue(ctx, job)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Enqueued add job: %s\n", id)
	}

	for i := 0; i < 3; i++ {
		b, _ := json.Marshal(Payload{Arg1: i + 2, Arg2: 3})
		job, err := tasqueue.NewJob("multiply", b, tasqueue.JobOpts{})
		if err != nil {
			log.Fatal(err)
		}
		id, err := srv.Enqueue(ctx, job)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Enqueued multiply job: %s\n", id)
	}

	// Failed jobs (with retries)
	for i := 0; i < 2; i++ {
		b, _ := json.Marshal(Payload{Arg1: i, Arg2: i})
		job, err := tasqueue.NewJob("fail", b, tasqueue.JobOpts{MaxRetries: 2})
		if err != nil {
			log.Fatal(err)
		}
		id, err := srv.Enqueue(ctx, job)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Enqueued fail job: %s\n", id)
	}

	// Create and enqueue a chain
	log.Println("\n=== Creating a chain ===")
	var chainJobs []tasqueue.Job
	for i := 0; i < 3; i++ {
		b, _ := json.Marshal(Payload{Arg1: i, Arg2: 1})
		job, err := tasqueue.NewJob("add", b, tasqueue.JobOpts{})
		if err != nil {
			log.Fatal(err)
		}
		chainJobs = append(chainJobs, job)
	}

	chain, err := tasqueue.NewChain(chainJobs, tasqueue.ChainOpts{})
	if err != nil {
		log.Fatal(err)
	}

	chainID, err := srv.EnqueueChain(ctx, chain)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Enqueued chain: %s\n", chainID)

	// Create and enqueue a group
	log.Println("\n=== Creating a group ===")
	var groupJobs []tasqueue.Job
	for i := 0; i < 4; i++ {
		b, _ := json.Marshal(Payload{Arg1: i * 2, Arg2: i * 3})
		job, err := tasqueue.NewJob("multiply", b, tasqueue.JobOpts{})
		if err != nil {
			log.Fatal(err)
		}
		groupJobs = append(groupJobs, job)
	}

	group, err := tasqueue.NewGroup(groupJobs, tasqueue.GroupOpts{})
	if err != nil {
		log.Fatal(err)
	}

	groupID, err := srv.EnqueueGroup(ctx, group)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Enqueued group: %s\n", groupID)

	log.Println("\n=== Worker running ===")
	log.Println("Jobs created:")
	log.Println("  - 5 add jobs (will succeed)")
	log.Println("  - 3 multiply jobs (will succeed)")
	log.Println("  - 2 fail jobs (will fail after retries)")
	log.Println("  - 1 chain of 3 jobs")
	log.Println("  - 1 group of 4 jobs")
	log.Println("\nView these jobs in the Tasqueue-UI dashboard!")
	log.Println("Press Ctrl+C to stop")

	// Keep running
	<-ctx.Done()
	log.Println("Shutting down...")
}
