package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/acim/test-consul-leader-election/pkg/cloud/consul"
	"github.com/robfig/cron"
)

func main() {
	me := os.Getenv("APP_MY_ID")
	sn := os.Getenv("SERVICE_NAME")
	consulAddress := "consul:8500"

	c := cron.New()
	c.AddFunc("0,30 * * * * *", func() {
		log.Printf("%s cron start\n", me)
		lock, err := consul.NewSingleRunLock(consulAddress, fmt.Sprintf("service/%s/lock/", sn))
		if err != nil {
			log.Printf("%s aborting job: %v", me, err)
			return
		}
		lead, err, unlock := lock.Lock()
		if err != nil {
			log.Printf("%s %v\n", me, err)
			return
		}
		if lead {
			defer unlock()
			log.Printf("%s working\n", me)
			time.Sleep(15 * time.Second)
			log.Printf("%s finished\n", me)
			return
		}
		log.Printf("%s gave up\n", me)
	})
	c.Start()
	defer c.Stop()

	select {}
}
