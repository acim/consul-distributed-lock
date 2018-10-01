// Package consul provides wrapper interface to Consul.
package consul

import (
	"context"
	"fmt"

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
	port    int
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
	return &client{
		consul:  c,
		svcName: serviceName,
		svcID:   serviceInstanceID,
		port:    port,
	}, nil
}

// Register implements Client interface.
func (c *client) Register() error {
	reg := &api.AgentServiceRegistration{
		ID:   c.svcID,
		Name: c.svcName,
		Port: c.port,
	}
	return c.consul.Agent().ServiceRegister(reg)
}

// Deregister implements Client interface.
func (c *client) Deregister(id string) error {
	return c.consul.Agent().ServiceDeregister(id)
}

// Service implements Client interface.
func (c *client) Service(service, tag string) ([]*api.ServiceEntry, *api.QueryMeta, error) {
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

func (c *client) Lock(ctx context.Context) (<-chan struct{}, error) {
	var err error
	c.lock, err = c.consul.LockKey(fmt.Sprintf("service/%s/lock/", c.svcName))
	if err != nil {
		return nil, fmt.Errorf("Lock: %v", err)
	}

	ch, err := c.lock.Lock(ctx.Done())
	if err != nil {
		return nil, fmt.Errorf("Lock: %v", err)
	}

	return ch, nil
}

func (c *client) Unlock() error {
	return c.lock.Unlock()
}
