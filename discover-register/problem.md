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
 etcd-test imports
	go.etcd.io/etcd/clientv3 tested by
	go.etcd.io/etcd/clientv3.test imports
	github.com/coreos/etcd/integration imports
	github.com/coreos/etcd/proxy/grpcproxy imports
	google.golang.org/grpc/naming: module google.golang.org/grpc@latest found (v1.51.0), but does not contain package google.golang.org/grpc/naming
```
解决方式  go mod edit -replace google.golang.org/grpc=google.golang.org/grpc@v1.26.0