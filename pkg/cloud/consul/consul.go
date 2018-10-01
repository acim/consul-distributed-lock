package consul

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
)

type singleRunProcess struct {
	client    *api.Client
	session   *api.Session
	sessionID string
	lock      *api.Lock
	lockKey   string
}

func NewSingleProcess(consulAddress string, connRetries int, retryDuration, lockWaitTime time.Duration, lockKey string) (*singleRunProcess, error) {
	config := api.DefaultConfig()
	config.Address = consulAddress
	var consul *api.Client
	var err error

	for i := 0; i < connRetries; i++ {
		consul, err = api.NewClient(config)
		if err == nil {
			break
		}
		time.Sleep(retryDuration)
		continue
	}

	if consul == nil {
		return nil, fmt.Errorf("error connecting: %s", consulAddress)
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
		Key:          lockKey,
		Session:      id,
		SessionName:  se.Name,
		LockWaitTime: lockWaitTime,
		LockTryOnce:  true,
	}
	lock, err := consul.LockOpts(opts)
	if err != nil {
		return nil, fmt.Errorf("error creating lock: %v", err)
	}

	return &singleRunProcess{
		client:    consul,
		session:   session,
		sessionID: id,
		lock:      lock,
		lockKey:   lockKey,
	}, nil
}

func (l *singleRunProcess) Lock() (bool, error, func() error) {
	leaderCh, err := l.lock.Lock(nil)
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

func (l *singleRunProcess) unlock() error {
	defer l.session.Destroy(l.sessionID, nil)

	err := l.lock.Unlock()
	if err != nil {
		return fmt.Errorf("error unlocking: %v", err)
	}

	l.client.KV().Delete(l.lockKey, nil)

	return nil
}
