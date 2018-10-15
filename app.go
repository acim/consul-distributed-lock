package main

import (
	"log"
	"os"
	"time"

	"github.com/acim/test-consul-leader-election/pkg/cloud/consul"
	"github.com/hashicorp/consul/api"
	"github.com/robfig/cron"
)

func main() {
	me := os.Getenv("APP_MY_ID")
	sn := os.Getenv("SERVICE_NAME")
	config := api.DefaultConfig()
	config.Address = "consul:8500"
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatal("can't initialize consul client")
	}

	c := cron.New()
	c.AddFunc("0,30 * * * * *", func() {
		log.Printf("%s cron start\n", me)
		lock, unlock, err := consul.NewSimpleLock(client, sn, 3*time.Second)
		if err != nil {
			log.Printf("%s error getting simple lock: %v", me, err)
			return
		}
		lead, err := lock()
		if err != nil {
			log.Printf("%s error locking %v\n", me, err)
			return
		}

		if !lead {
			log.Printf("%s gave up\n", me)
			return
		}

		log.Printf("%s working\n", me)
		time.Sleep(15 * time.Second) // something useful to do
		log.Printf("%s finished\n", me)

		err = unlock()
		if err != nil {
			log.Printf("%s error unlocking %v\n", me, err)
			return
		}
	})
	c.Start()
	defer c.Stop()

	select {}
}
