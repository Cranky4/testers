package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"sync"
	"syscall"

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

func main() {
	conn := getConn()
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	count := 10
	queues := make([]amqp.Queue, count)
	for i := 0; i < count; i++ {
		q, err := ch.QueueDeclare(
			fmt.Sprintf("hello-%d", i), // name
			false,                      // durable
			false,                      // delete when unused
			false,                      // exclusive
			false,                      // no-wait
			nil,                        // arguments
		)
		failOnError(err, "Failed to declare a queue")

		queues[i] = q
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	wg := &sync.WaitGroup{}
	for _, v := range queues {
		wg.Add(1)

		go func(ctx context.Context, queue amqp.Queue, wg *sync.WaitGroup) {
			defer wg.Done()

			msgs, err := ch.Consume(
				queue.Name, // queue
				"",         // consumer
				true,       // auto-ack
				false,      // exclusive
				false,      // no-local
				false,      // no-wait
				nil,        // args
			)
			failOnError(err, "Failed to register a consumer")

			for {
				select {
				case d := <-msgs:
					log.Printf("Received a message from %s: %s", queue.Name, d.Body)
				case <-ctx.Done():
					return
				}
			}
		}(ctx, v, wg)
	}

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	wg.Wait()
}
