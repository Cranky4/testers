package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func getConn() *amqp.Connection {
	conn, err := amqp.Dial("amqp://...")
	failOnError(err, "Failed to connect to RabbitMQ")

	return conn
}

type QTicker struct {
	Q      amqp.Queue
	Ticker *time.Ticker
}

func main() {
	conn := getConn()
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	count := 10
	queues := make([]QTicker, count)
	for i := 0; i < count; i++ {
		ticker := time.NewTicker(10 * time.Millisecond)

		q, err := ch.QueueDeclare(
			fmt.Sprintf("hello-%d", i), // name
			false,                      // durable
			false,                      // delete when unused
			false,                      // exclusive
			false,                      // no-wait
			nil,                        // arguments
		)
		failOnError(err, "Failed to declare a queue")

		queues[i] = QTicker{
			Q:      q,
			Ticker: ticker,
		}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	wg := &sync.WaitGroup{}

	for i, v := range queues {
		wg.Add(1)
		go func(ctx context.Context, v QTicker, num int, wg *sync.WaitGroup) {
			defer wg.Done()
			defer v.Ticker.Stop()

			for {
				select {
				case <-v.Ticker.C:
					body := "Hello World!"
					err = ch.PublishWithContext(ctx,
						"",       // exchange
						v.Q.Name, // routing key
						false,    // mandatory
						false,    // immediate
						amqp.Publishing{
							ContentType: "text/plain",
							Body:        []byte(body),
						})
					failOnError(err, "Failed to publish a message")

					log.Printf(" [x] Sent to queue %s, body %s\n", v.Q.Name, body)
				case <-ctx.Done():
					return
				}
			}
		}(ctx, v, i, wg)
	}

	wg.Wait()

	log.Printf("DONE!")
}
