package DHT

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	PUT    = EventOp(2)
	DELETE = EventOp(3)
)

type Server interface {
	Register(seedAddr string) error
	Revoke(string) error
	Watch(string) Watcher
	KeepAlive(ctx context.Context, duration time.Duration)
	GetGroup(string) (RangeResponse, error)
	Listen() error
	Close() error
	CheckNodeStatus(string)
}
type Watcher <-chan WatchResponse
type Event struct {
	Type EventOp
	Kv   KeyValue
}
type EventOp int32
type WatchResponse struct {
	Header ResponseHeader
	Events []*Event
}
type RangeResponse struct {
	Header ResponseHeader
	Kvs    []KeyValue
}
type ResponseHeader struct {
	ClusterId uint64
	MemberId  uint64
	Revision  int64
	RaftTerm  uint64
}
type KeyValue struct {
	Key   []byte
	Value []byte
}

type NetServer struct {
	server    *dhtNetWorker
	watchChan map[string]chan WatchResponse
	seedAddr  string
}

// 解析传入的对端节点信息
func (d *NetServer) parseSeed() (result []net.Addr, err error) {
	for _, url := range strings.Split(d.seedAddr, ",") {
		urls := strings.Split(url, "://")
		if len(urls) != 2 {
			continue
		}
		netType := urls[0]
		a := strings.Split(urls[1], ":")
		ip := a[0]
		port, e := strconv.Atoi(a[1])
		if e != nil {
			continue
		}
		switch netType {
		case "udp":
			result = append(result, &net.UDPAddr{IP: net.ParseIP(ip), Port: port})
		case "tcp":
			result = append(result, &net.TCPAddr{IP: net.ParseIP(ip), Port: port})
		default:
			return nil, errors.New("未知的通讯类型")
		}
	}
	return
}
func (d *NetServer) Register(seedAddr string) error {
	var (
		n   *node
		e   error
		ers []error
	)
	d.seedAddr = seedAddr
	seedList, err := d.parseSeed()
	if err != nil {
		return err
	}
	for _, addr := range seedList {
		n, e = d.server.register(addr)
		if e == nil {
			d.server.table.append(n)
			// 有一个节点连接上了就可以了
			d.server.getNodeTable(addr)
			return nil
		} else {
			ers = append(ers, e)
		}
	}

	return e
}
func (d *NetServer) Revoke(nodeId string) error {
	return nil
}
func (d *NetServer) GetGroup(groupName string) (RangeResponse, error) {
	var result []KeyValue
	key := []byte(groupName)
	for _, n := range d.server.table.get(groupName) {
		fmt.Printf("从%s中获取的node有:%s\n", groupName, n.grpcServer)
		result = append(result, KeyValue{
			Key:   key,
			Value: []byte(n.grpcServer),
		})
	}
	return RangeResponse{Kvs: result}, nil
}
func (d *NetServer) Watch(nodeType string) Watcher {
	closeCh := make(chan WatchResponse, 1)
	d.watchChan[nodeType] = closeCh
	if d.server.watch == nil {
		nodeChan := make(chan node, 5)
		d.server.watch = nodeChan
		go func() {
			for n := range d.server.watch {
				fmt.Printf("节点变化消息: %s 节点，状态：%t\n", n.grpcServer, n.status)
				ch, ok := d.watchChan[n.ServerType]
				if !ok {
					continue
				}
				kv := KeyValue{
					Key:   []byte(n.ServerType),
					Value: []byte(n.grpcServer),
				}
				var e EventOp
				if n.status {
					e = PUT
				} else {
					e = DELETE
				}
				ch <- WatchResponse{
					Events: []*Event{{Type: e, Kv: kv}},
				}
			}
		}()
	}

	return closeCh
}
func (d *NetServer) KeepAlive(ctx context.Context, duration time.Duration) {

}
func (d *NetServer) Listen() error {
	return d.server.listen()
}
func (d *NetServer) Close() error {
	d.server.offline(d.server.selfNode.ServerType, d.server.selfNode.id)
	d.server.close()
	return nil
}
func (d *NetServer) CheckNodeStatus(groupName string) {
	d.server.checkStatus(d.server.table.get(groupName))
}
func (d *NetServer) FindNetWorker() {
	var (
		addrs []net.Addr
		err   error
	)
	if d.server.table.length() == 0 {
		addrs, err = d.parseSeed()
		if err != nil {
			return
		}
	} else {
		for _, n := range d.server.table.nodeTable {
			addrs = append(addrs, n.addr)
		}
	}

	for _, addr := range addrs {
		n, e := d.server.register(addr)
		if e == nil {
			d.server.table.append(n)
			// 有一个节点连接上了就可以了
			d.server.getNodeTable(addr)
		}
	}
}
func NewNetServer(port int, serverIp, serverType, serverName, grpcAddress string) (Server, error) {
	server, err := newDhtServer(port, serverIp, serverType, serverName, grpcAddress)
	if err != nil {
		return nil, err
	}
	s := NetServer{server: server, watchChan: make(map[string]chan WatchResponse)}
	return &s, nil
}
