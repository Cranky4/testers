package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func ExampleClient() {
	rdbW := redis.NewClient(&redis.Options{
		Addr:     "",
		Password: "",
		DB:       0, // use default DB
	})

	rdbR := redis.NewClient(&redis.Options{
		Addr:     "",
		Password: "",
		DB:       0, // use default DB
	})

	fmt.Println("start")
	err := rdbW.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second)
	val, err := rdbR.Get(ctx, "key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)

	val2, err := rdbR.Get(ctx, "key2").Result()
	if err == redis.Nil {
		fmt.Println("key2 does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key2", val2)
	}
	// Output: key value
	// key2 does not exist

	fmt.Println("done")
}

func main() {
	ExampleClient()
}
