package consul

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
)

// NewSimpleLock creates *SimpleLock.
func NewSimpleLock(consul *api.Client, lockKey string, lockWaitTime time.Duration) (func() (bool, error), func() error, error) {
	var err error
	sl := &simpleLock{
		consul:  consul,
		lockKey: lockKey,
	}

	sl.session = sl.consul.Session()

	se := &api.SessionEntry{
		Name:     api.DefaultLockSessionName,
		Behavior: api.SessionBehaviorDelete,
	}
	sl.sessionID, _, err = sl.session.CreateNoChecks(se, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating session: %v", err)
	}

	opts := &api.LockOptions{
		Key:          sl.lockKey,
		Session:      sl.sessionID,
		SessionName:  se.Name,
		LockWaitTime: lockWaitTime,
		LockTryOnce:  true,
	}
	sl.lock, err = sl.consul.LockOpts(opts)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating lock: %v", err)
	}

	return sl.l, sl.u, err
}

type simpleLock struct {
	consul    *api.Client
	session   *api.Session
	sessionID string
	lock      *api.Lock
	lockKey   string
	locked    bool
}

func (sl *simpleLock) l() (bool, error) {
	leaderCh, err := sl.lock.Lock(nil)
	if err != nil {
		return false, fmt.Errorf("error locking: %v", err)
	}
	if leaderCh == nil {
		return false, nil
	}
	select {
	case <-leaderCh:
		return false, fmt.Errorf("should be leader, unexpected error: %v", err)
	default:
	}

	sl.locked = true
	return true, nil
}

func (sl *simpleLock) u() error {
	if !sl.locked {
		return nil
	}
	defer sl.session.Destroy(sl.sessionID, nil)

	err := sl.lock.Unlock()
	if err != nil {
		return fmt.Errorf("error unlocking: %v", err)
	}

	sl.consul.KV().Delete(sl.lockKey, nil)

	return nil
}
