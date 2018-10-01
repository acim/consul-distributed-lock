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
		log.Printf("%s cron start", me)
		defer log.Printf("%s cron end", me)
		opts := &api.LockOptions{
			Key:        fmt.Sprintf("service/%s/lock/", sn),
			Value:      []byte(fmt.Sprintf("set by %s", me)),
			SessionTTL: "60s",
		}

		lock, err := consul.LockOpts(opts)
		if err != nil {
			log.Printf("%s LockOpts: %v", me, err)
			return
		}

		haveLock := false

		stop := make(chan struct{})
		time.AfterFunc(5*time.Second, func() {
			if !haveLock {
				close(stop)
			}
		})

		lead := make(chan struct{})
		go func() {
			log.Printf("%s before lock", me)
			_, err = lock.Lock(stop)
			log.Printf("%s after lock", me)
			if err != nil {
				log.Printf("%s Lock: %v", me, err)
				return
			}
			select {
			case <-stop:
				return
			default:
				log.Printf("%s lock", me)
				haveLock = true
				close(lead)
			}

		}()

		select {
		case <-lead:
			log.Printf("%s working", me)
			time.Sleep(15 * time.Second)
			log.Printf("%s unlock", me)
			err = lock.Unlock()
			if err != nil {
				log.Printf("%s unlock: %v", me, err)
			}
		case <-stop:
			log.Printf("%s aborting", me)
		}
	})
	c.Start()
	defer c.Stop()

	select {}
}
