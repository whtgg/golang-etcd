package main

import (
	"context"
	"go.etcd.io/etcd/clientv3"
	"log"
	"time"
)

// ServiceRegister 创建租约注册服务
type ServiceRegister struct {
	client        *clientv3.Client
	leaseId       clientv3.LeaseID
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	key           string
	val           string
	ctx           context.Context
}

// NewServiceRegister 新建注册服务
func NewServiceRegister(endpoints []string, key, val string, ttl int64) (*ServiceRegister, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	srv := &ServiceRegister{
		client: client,
		key:    key,
		val:    val,
		ctx:    context.Background(),
	}

	if err = srv.putKeyWithLease(ttl); err != nil {
		return nil, err
	}

	return srv, nil
}

func (srv *ServiceRegister) putKeyWithLease(ttl int64) error {
	// 设置租约时间
	leaseResp, err := srv.client.Grant(srv.ctx, ttl)
	if err != nil {
		return err
	}
	_, err = srv.client.Put(srv.ctx, srv.key, srv.val, clientv3.WithLease(clientv3.LeaseID(leaseResp.ID)))
	if err != nil {
		return err
	}

	keepAliveRespChan, err := srv.client.KeepAlive(srv.ctx, clientv3.LeaseID(leaseResp.ID))
	if err != nil {
		return err
	}

	srv.leaseId = clientv3.LeaseID(leaseResp.ID)
	log.Println("releaseId--", leaseResp.ID)
	srv.keepAliveChan = keepAliveRespChan
	return nil
}

// 监听续租
func (srv *ServiceRegister) listenLeaseRespChan() {
	for leaseKeepResp := range srv.keepAliveChan {
		log.Println("续约成功", leaseKeepResp)
	}
	log.Println("续约失败")
}

// 注销服务
func (srv *ServiceRegister) close() error {
	if _, err := srv.client.Revoke(srv.ctx, srv.leaseId); err != nil {
		return err
	}
	log.Println("撤销服务")
	return srv.client.Close()
}

func main() {
	var endpoints = []string{"localhost:2379"}
	srv, err := NewServiceRegister(endpoints, "/node1", "localhost:8000", 5)
	if err != nil {
		log.Fatal(err)
	}
	go srv.listenLeaseRespChan()
	select {
	case <-time.After(time.Minute * 5):
		_ = srv.close()
	}
}
