package lock

import (
	"context"
	"fmt"
	v3 "github.com/coreos/etcd/clientv3"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	"log"
	"time"
)

func AcquireLock() (bool, error) {
	// 连接Etcd
	config := clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	}

	client, err := clientv3.New(config)
	if err != nil {
		log.Println(err)
		return false, err
	}
	// 创建一个租约
	lease := clientv3.NewLease(client)
	leaseResp, err := lease.Grant(context.TODO(), 5)
	if err != nil {
		log.Println(err)
		return false, err
	}
	leaseId := leaseResp.ID
	// 释放租约
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	defer lease.Revoke(context.TODO(), leaseId)

	leaseChan, err := lease.KeepAlive(ctx, leaseId)
	if err != nil {
		log.Println(err)
		return false, err
	}

	// 续约监听
	go func() {
		for {
			select {
			case keepResp := <-leaseChan:
				if leaseChan == nil {
					log.Println("租约已经失效")
					goto END
				} else {
					log.Println("收到自动续约应答", keepResp.ID)
				}
			}
		}
	END:
	}()

	kv := clientv3.NewKV(client)
	txn := kv.Txn(context.TODO())

	txn.If(clientv3.Compare(clientv3.CreateRevision("lock"), "=", 0)).
		Then(clientv3.OpPut("lock", "g", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet("lock"))

	txnResp, err := txn.Commit()
	if err != nil {
		log.Println(err)
		return false, err
	}

	if !txnResp.Succeeded {
		fmt.Println("锁被占用:", string(txnResp.Responses[0].GetResponseRange().Kvs[0].Value))
	}
	log.Println(err)
	time.Sleep(5 * time.Second)
	return true, nil
}

func AcquireLockThird() {
	cli, err := v3.New(v3.Config{Endpoints: []string{"localhost:2379"}})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	// create two separate sessions for lock competition
	s1, err := concurrency.NewSession(cli)
	if err != nil {
		log.Fatal(err)
	}
	defer s1.Close()
	m1 := concurrency.NewMutex(s1, "/my-lock/")

	s2, err := concurrency.NewSession(cli)
	if err != nil {
		log.Fatal(err)
	}
	defer s2.Close()
	m2 := concurrency.NewMutex(s2, "/my-lock/")

	// acquire lock for s1
	if err := m1.Lock(context.TODO()); err != nil {
		log.Fatal(err)
	}
	fmt.Println("acquired lock for s1")

	m2Locked := make(chan struct{})
	go func() {
		defer close(m2Locked)
		// wait until s1 is locks /my-lock/
		if err := m2.Lock(context.TODO()); err != nil {
			log.Fatal("-------", err)
		}
	}()

	if err := m1.Unlock(context.TODO()); err != nil {
		log.Fatal(err)
	}
	fmt.Println("released lock for s1")

	<-m2Locked
	fmt.Println("acquired lock for s2")
}
