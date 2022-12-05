package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ServiceDiscover 服务发现
type ServiceDiscover struct {
	client  *clientv3.Client
	servers map[string]string //服务列表
	lock    sync.Mutex
	//. ctx     context.Context
}

// NewServiceDiscover 初始化
func NewServiceDiscover(endpoints []string) (*ServiceDiscover, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &ServiceDiscover{
		client:  client,
		servers: make(map[string]string),
	}, nil
}

// WatchService 初始化服务列表和监视
func (sd *ServiceDiscover) WatchService(key string) error {
	getResp, err := sd.client.Get(context.Background(), key, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	fmt.Println(getResp)
	return nil

}

// watcher 监听前缀
func (sd *ServiceDiscover) watcher(prefix string) {
	watchChan := sd.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	log.Printf("watching prefix %s now...", prefix)
	for ch := range watchChan {
		for _, ev := range ch.Events {
			switch ev.Type {
			case mvccpb.PUT: //修改或者新增
				sd.SetServiceList(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE: //删除
				sd.DelServiceList(string(ev.Kv.Key))
			}
		}
	}
}

// SetServiceList 新增服务地址
func (sd *ServiceDiscover) SetServiceList(key, val string) {
	sd.lock.Lock()
	defer sd.lock.Unlock()
	sd.servers[key] = val
	log.Println("pub key:", key, "val:", val)
}

// DelServiceList 删除服务地址
func (sd *ServiceDiscover) DelServiceList(key string) {
	sd.lock.Lock()
	defer sd.lock.Unlock()
	delete(sd.servers, key)
	log.Println("del key:", key)
}

// GetServices 获取服务地址
func (sd *ServiceDiscover) GetServices() []string {
	sd.lock.Lock()
	defer sd.lock.Unlock()
	adders := make([]string, 0)

	for _, v := range sd.servers {
		adders = append(adders, v)
	}
	return adders
}

// Close 关闭服务
func (sd *ServiceDiscover) Close() error {
	return sd.client.Close()
}
func main() {
	var endpoints = []string{"localhost:2379"}
	ser, err := NewServiceDiscover(endpoints)
	if err != nil {
		log.Fatal("NewServiceDiscover", err)
	}
	defer ser.Close()
	if err := ser.WatchService("/web/"); err != nil {
		log.Fatal("WatchService", err)
	}

	// 监控系统信号，等待 ctrl + c 系统信号通知服务关闭
	c := make(chan os.Signal, 1)
	go func() {
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	}()

	for {
		select {
		case <-time.Tick(10 * time.Second):
			log.Println(ser.GetServices())
		case <-c:
			log.Println("server discovery exit")
			return
		}
	}
}
