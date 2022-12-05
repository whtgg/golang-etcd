# golang-etcd

### Etcd特性
* 简单：etcd 的安装简单，且为用户提供了 HTTP API，使用起来也很简单。
* 存储：etcd 的基本功能，数据分层存储在文件目录中，类似于我们日常使用的文件系统。
* Watch 机制：Watch 指定的键、前缀目录的更改，并对更改时间进行通知。
* 安全通信：支持 SSL 证书验证。
* 高性能：etcd 单实例可以支持 2K/s 读操作，每个实例1000次写入/秒，官方也有提供基准测试脚本。
* 一致可靠：基于 Raft 共识算法，实现分布式系统内部数据存储、服务调用的一致性和高可用性。

### Etcd使用场景
* 键值对存储
> etcd 是一个用于键值存储的组件，存储是 etcd 最基本的功能，其他应用场景都建立在 etcd 的可靠存储上。比如 Kubernetes 将一些元数据存储在 etcd 中，将存储状态数据的复杂工作交给 etcd，Kubernetes 自身的功能和架构就能更加稳定。
etcd 基于 Raft 算法，能够有力地保证分布式场景中的一致性。各个服务启动时注册到 etcd 上，同时为这些服务配置键的 TTL 时间。注册到 etcd 上面的各个服务实例通过心跳的方式定期续租，实现服务实例的状态监控。
* 消息发布与订阅
> 通过构建 etcd 消息中间件，服务提供者发布对应主题的消息，消费者则订阅他们关心的主题，一旦对应的主题有消息发布，就会产生订阅事件，消息中间件就会通知该主题所有的订阅者。
* 分布式锁 
> 分布式系统中涉及多个服务实例，存在跨进程之间资源调用，对于资源的协调分配，单体架构中的锁已经无法满足需要，需要引入分布式锁的概念。etcd 基于 Raft 算法，实现分布式集群的一致性，存储到 etcd 集群中的值必然是全局一致的，因此基于 etcd 很容易实现分布式锁。
* 集群监控与 Leader 竞选
> 通过watcher机制可以第一时间发现节点变动
> 节点可以设置TTL key，比如每隔 30s 发送一次心跳使代表该机器存活的节点继续存在，否则节点消失
* 负载均衡
> 此处指的负载均衡均为软负载均衡，分布式系统中，为了保证服务的高可用以及数据的一致性，通常都会把数据和服务部署多份，以此达到对等服务，即使其中的某一个服务失效了，也不影响使用。由此带来的坏处是数据写入性能下降，而好处则是数据访问时的负载均衡。因为每个对等服务节点上都存有完整的数据，所以用户的访问流量就可以分流到不同的机器上

### Etcd基本架构
> etcd 有 etcd Server、gRPC Server、存储相关的 MVCC 、Snapshot、WAL，以及 Raft 模块。
* etcd Server 用于对外接收和处理客户端的请求
* gRPC Server 则是 etcd 与其他 etcd 节点之间的通信和信息同步
* MVCC，即多版本控制，etcd 的存储模块，键值对的每一次操作行为都会被记录存储，这些数据底层存储在 BoltDB 数据库中
* WAL，预写式日志，etcd 中的数据提交前都会记录到日志
* Snapshot 快照，以防 WAL 日志过多，用于存储某一时刻 etcd 的所有数据。Snapshot 和 WAL 相结合，etcd 可以有效地进行数据存储和节点故障恢复等操作。

#### Etcd和Redis的区别
etcd和redis都支持键值存储，也支持分布式特性，redis支持的数据格式更加丰富，但是他们两个定位和应用场景不一样，关键差异如下：
* redis在分布式环境下不是强一致性的，可能会丢失数据，或者读取不到最新数据
* redis的数据变化监听机制没有etcd完善
* etcd强一致性保证数据可靠性，导致性能上要低于redis
* etcd和ZooKeeper是定位类似的项目，跟redis定位不一样

### Etcd和ZooKeeper区别
ZooKeeper的不足
* 复杂：ZooKeeper的部署维护复杂，管理员需要掌握一系列的知识和技能；而 Paxos 强一致性算法也是素来以复杂难懂而闻名于世；另外，ZooKeeper的使用也比较复杂，需要安装客户端。
* 难以维护: 引入大量的依赖 维护起来容易出错
* 发展缓慢
Etcd改进
* 简单 
* 数据持久化
* 安全

### 部署 
docker-compose 部署

### 代码结构
* discover-register 服务注册和发现
* distributed-lock  分布式锁
> 分布式与单机环境最大的不同在于它不是多线程而是多进程。由于多线程可以共享堆内存，因此可以简单地采取内存作为标记存储位置。而多进程可能都不在同一台物理机上，就需要将标记存储在一个所有进程都能看到的地方。

### 问题
```
go: etcd-test imports
	go.etcd.io/etcd/clientv3 tested by
	go.etcd.io/etcd/clientv3.test imports
	github.com/coreos/etcd/auth imports
	github.com/coreos/etcd/mvcc/backend imports
	github.com/coreos/bbolt: github.com/coreos/bbolt@v1.3.6: parsing go.mod:
	module declares its path as: go.etcd.io/bbolt
	        but was required as: github.com/coreos/bbolt
```
解决方式 go mod edit -replace github.com/coreos/bbolt=go.etcd.io/bbolt@v1.3.4

```
 undefined: resolver.BuildOption undefined: resolver.ResolveNowOption 
 或者
 etcd-test imports
	go.etcd.io/etcd/clientv3 tested by
	go.etcd.io/etcd/clientv3.test imports
	github.com/coreos/etcd/integration imports
	github.com/coreos/etcd/proxy/grpcproxy imports
	google.golang.org/grpc/naming: module google.golang.org/grpc@latest found (v1.51.0), but does not contain package google.golang.org/grpc/naming
```
解决方式  go mod edit -replace google.golang.org/grpc=google.golang.org/grpc@v1.26.0

