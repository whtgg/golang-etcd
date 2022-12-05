package lock_test

import (
	lock "go-etcd/distributed-lock"
	"testing"
)

func TestAcquireLock(t *testing.T) {
	t.Run("lock", func(t *testing.T) {
		_, _ = lock.AcquireLock()
	})
}

func TestAcquireLockThird(t *testing.T) {
	lock.AcquireLockThird()
}
