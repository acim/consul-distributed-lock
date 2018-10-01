// Package consul provides wrapper interface to Consul.
package consul

import (
	"fmt"
	"log"

	"github.com/hashicorp/consul/api"
)

// https://medium.com/@mthenw/distributed-locks-with-consul-and-golang-c4eccc217dd5
// https://github.com/dmitriyGarden/consul-leader-election
// https://github.com/docker/leadership
// https://github.com/Comcast/go-leaderelection
// https://www.consul.io/docs/guides/leader-election.html
// https://zookeeper.apache.org/doc/current/recipes.html#sc_leaderElection

//Client provides an interface for communicating with Consul.
// type Client interface {
// 	// Get a Service from Consul.
// 	Service(string, string) ([]string, error)
// 	// Register a service with Consul.
// 	Register(string, int) error
// 	// Deregister a service with Consul.
// 	Deregister(string) error
// }

type Client struct {
	consul  *api.Client
	svcName string
	svcID   string
	lock    *api.Lock
}

// NewClient returns a Client interface for given consul address.
func NewClient(consulAddr, serviceName, serviceInstanceID string, port int) (*Client, error) {
	config := api.DefaultConfig()
	config.Address = consulAddr
	c, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Client{
		consul:  c,
		svcName: serviceName,
		svcID:   serviceInstanceID,
	}, nil
}

// Register implements Client interface.
func (c *Client) Register() error {
	reg := &api.AgentServiceRegistration{
		ID:   c.svcID,
		Name: c.svcName,
	}
	return c.consul.Agent().ServiceRegister(reg)
}

// Deregister implements Client interface.
func (c *Client) Deregister(id string) error {
	return c.consul.Agent().ServiceDeregister(id)
}

// Service implements Client interface.
func (c *Client) Service(service, tag string) ([]*api.ServiceEntry, *api.QueryMeta, error) {
	passingOnly := true
	addrs, meta, err := c.consul.Health().Service(service, tag, passingOnly, nil)
	if len(addrs) == 0 && err == nil {
		return nil, nil, fmt.Errorf("service ( %s ) was not found", service)
	}
	if err != nil {
		return nil, nil, err
	}
	return addrs, meta, nil
}

func (c *Client) Lock() bool {
	// (string, *WriteMeta, error)
	x, y, err := c.consul.Session().CreateNoChecks(&api.SessionEntry{
		ID:   c.svcID,
		Name: c.svcName,
	}, nil)

	log.Println(x, y, err)

	// var err error
	// c.lock, err = c.consul.LockKey(fmt.Sprintf("service/%s/lock/", c.svcName))
	// if err != nil {
	// 	log.Printf("Lock: %v", err)
	// }

	// stop := make(chan struct{})
	// time.AfterFunc(500*time.Millisecond, func() {
	// 	stop <- struct{}{}
	// })

	// log.Println(c.svcID, ">>>", 1)
	// ch, err := c.lock.Lock(stop)
	// if err != nil {
	// 	log.Printf("Lock: %v", err)
	// }
	// log.Println(c.svcID, ">>>", 2)

	// select {
	// case <-ch:
	// 	log.Println(c.svcID, ">>>", 3)
	// 	return false
	// default:
	// 	log.Println(c.svcID, ">>>", 4)
	// 	return true
	// }
	return false
}

func (c *Client) Unlock() error {
	return c.lock.Unlock()
}
