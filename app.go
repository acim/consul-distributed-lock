package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/robfig/cron"
)

func main() {
	me := os.Getenv("APP_MY_ID")
	sn := os.Getenv("SERVICE_NAME")
	config := api.DefaultConfig()
	config.Address = "consul:8500"
	var consul *api.Client
	var err error
	for {
		consul, err = api.NewClient(config)
		if err != nil {
			log.Printf("%s can't connect to consul: %v\n", me, err)
			time.Sleep(3 * time.Second)
			continue
		}
		break
	}

	c := cron.New()
	c.AddFunc("0,30 * * * * *", func() {
		log.Printf("%s cron trigger", me)
		opts := &api.LockOptions{
			Key:        fmt.Sprintf("service/%s/lock/", sn),
			Value:      []byte(fmt.Sprintf("set by %s", me)),
			SessionTTL: "10s",
		}

		lock, err := consul.LockOpts(opts)
		if err != nil {
			log.Printf("%s Lock: %v", me, err)
			return
		}

		_, err = lock.Lock(nil)
		if err != nil {
			log.Printf("%s Lock: %v", me, err)
			return
		}
		log.Printf("%s lock, doing something", me)
		time.Sleep(20 * time.Second)
		log.Printf("%s unlock", me)
		err = lock.Unlock()
		if err != nil {
			log.Printf("%s unlock: %v", me, err)
		}

		// <-lockCh
		// log.Println(me, 2)
	})
	c.Start()
	defer c.Stop()

	select {}
}
