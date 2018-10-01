package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/acim/test-consul-leader-election/pkg/cloud/consul"
	"github.com/robfig/cron"
)

func main() {
	me := os.Getenv("APP_MY_ID")
	sn := os.Getenv("SERVICE_NAME")
	var client *consul.Client
	var err error
	for {
		client, err = consul.NewClient("consul:8500", sn, me, 0)
		if err != nil {
			log.Printf("%s can't connect to consul: %v\n", me, err)
			time.Sleep(3 * time.Second)
			continue
		}
		break
	}

	err = client.Register()
	if err != nil {
		log.Printf("%s can't register to consul: %v\n", me, err)
	}

	c := cron.New()
	c.AddFunc("0 * * * * *", func() {
		log.Printf("%s cron triggered at %s", me, time.Now().String())
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		lost, err := client.Lock(ctx)
		if err != nil {
			log.Printf("%s error acquiring lock: %v\n", me, err)
			return
		}

		select {
		case <-lost:
			log.Printf("%s didn't acquire lock", me)
			return
		default:
			log.Printf("%s acquired lock - doing something for 10 seconds", me)
			time.Sleep(10 * time.Second)
			err = client.Unlock()
			if err != nil {
				log.Printf("%s can't release lock: %v\n", me, err)
			}
		}
	})
	c.Start()
	defer c.Stop()

	select {}
}
