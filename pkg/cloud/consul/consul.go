package consul

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
)

type singleRunLock struct {
	c  *api.Client
	s  *api.Session
	id string
	l  *api.Lock
	k  string
}

func NewSingleRunLock(address, key string) (*singleRunLock, error) {
	config := api.DefaultConfig()
	config.Address = address
	var consul *api.Client
	var err error

	for i := 0; i < 3; i++ {
		consul, err = api.NewClient(config)
		if err == nil {
			break
		}
		time.Sleep(3 * time.Second)
		continue
	}

	if consul == nil {
		return nil, fmt.Errorf("error connecting: %s", address)
	}

	session := consul.Session()

	se := &api.SessionEntry{
		Name:     api.DefaultLockSessionName,
		Behavior: api.SessionBehaviorDelete,
	}
	id, _, err := session.CreateNoChecks(se, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating session: %v", err)
	}

	opts := &api.LockOptions{
		Key:          key,
		Session:      id,
		SessionName:  se.Name,
		LockWaitTime: time.Second,
		LockTryOnce:  true,
	}
	lock, err := consul.LockOpts(opts)
	if err != nil {
		return nil, fmt.Errorf("error creating lock: %v", err)
	}

	return &singleRunLock{
		c:  consul,
		s:  session,
		id: id,
		l:  lock,
		k:  key,
	}, nil
}

func (l *singleRunLock) Lock() (bool, error, func() error) {
	leaderCh, err := l.l.Lock(nil)
	if err != nil {
		return false, fmt.Errorf("error locking: %v", err), nil
	}
	if leaderCh == nil {
		return false, nil, nil
	}
	select {
	case <-leaderCh:
		return false, fmt.Errorf("error should be leader: %v", err), nil
	default:
	}

	return true, nil, l.unlock
}

func (l *singleRunLock) unlock() error {
	time.Sleep(10 * time.Second)
	defer l.s.Destroy(l.id, nil)

	err := l.l.Unlock()
	if err != nil {
		return fmt.Errorf("error unlocking: %v", err)
	}

	return nil
}
