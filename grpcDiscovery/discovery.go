package grpcDiscovery

// git@github.com:Bingjian-Zhu/etcd-example.git
import (
	"fmt"
	"google.golang.org/grpc/resolver"
	"log"
	"sync"
	"time"
	"videoDynamicAcquisition/grpcDiscovery/DHT"
)

// grpc 服务发现名称 在创建客户端的时候使用,（只能使用小写，大写会出现无法发现服务，因为注册的时候会强制转小写）
const (
	schema           = "dht"
	prefixServerName = "name"
)

//ServiceDiscovery 服务发现
type ServiceDiscovery struct {
	cli              DHT.Server //etcd client
	cc               resolver.ClientConn
	serverList       sync.Map //服务列表
	groupName        string   //监视的服务名称
	prefixServerName string   // 同一组中不同的服务器名称
	lock             *TimeLock
}

type ServerConfig struct {
	ServerType       string
	ServerName       string
	SeedAddr         string
	AwaitRegister    bool //  是否等待连接到网络
	NodePort         int
	ServerIp         string
	GrpcSeverAddress string
}

func init() {
	resolver.Register(&ServiceDiscovery{})
}

//NewServiceDiscovery  新建发现服务
func NewServiceDiscovery(config *ServerConfig) (*ServiceDiscovery, error) {
	ds, err := DHT.NewNetServer(config.NodePort, config.ServerIp, config.ServerType, config.ServerName, config.GrpcSeverAddress)
	if err != nil {
		return nil, err
	}
	go func() {
		err := ds.Listen()
		if err != nil {
			fmt.Printf("dht网络监听出现错误：%s", err.Error())
		}
	}()
	if config.AwaitRegister {
		//   如果种子节点未上线，那就定去循环请求，并把这个状态写到日志当中
		var er error
		er = ds.Register(config.SeedAddr)
		for er != nil {
			fmt.Printf("加入dht网络错误，等待重试.错误详情%s", er.Error())
			time.Sleep(time.Minute * 5)
			er = ds.Register(config.SeedAddr)
		}
	} else {
		go func() {
			//   如果种子节点未上线，那就定去循环请求，并把这个状态写到日志当中
			var er error
			er = ds.Register(config.SeedAddr)
			for er != nil {
				fmt.Printf("加入dht网络错误，等待重试.错误详情%s", er.Error())
				time.Sleep(time.Minute * 5)
				er = ds.Register(config.SeedAddr)
			}
		}()
	}
	return &ServiceDiscovery{
		cli: ds,
		lock: &TimeLock{
			TimeTick: 300,
		},
	}, nil
}

//Build 为给定目标创建一个新的`resolver`，当调用`grpc.Dial()`时执行
func (s *ServiceDiscovery) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	s.cc = cc
	prefixServer := target.URL.Query().Get(prefixServerName)
	s.groupName = target.URL.Path
	var (
		resp      DHT.RangeResponse
		err       error
		getServer string
	)
	getServer = s.groupName
	if prefixServer != "" {
		s.prefixServerName = prefixServer
		getServer = prefixServer
	}
	//根据服务名获取现有的节点
	resp, err = s.cli.GetGroup(getServer)
	if err != nil {
		return nil, err
	}

	for _, ev := range resp.Kvs {
		s.SetServiceList(string(ev.Key), string(ev.Value))
	}
	s.cc.UpdateState(resolver.State{Addresses: s.getServices()})
	//监视前缀，修改变更的server
	go s.watcher()
	return s, nil
}

// ResolveNow 监视目标更新
func (s *ServiceDiscovery) ResolveNow(rn resolver.ResolveNowOptions) {
	log.Printf("%s组当中有节点失效", s.groupName)
	var getServer string
	getServer = s.groupName
	if s.prefixServerName != "" {
		getServer = s.prefixServerName
	}
	if s.lock.LockByGroupName(getServer) {
		// 获取是否可执行
		s.cli.CheckNodeStatus(getServer)
		s.lock.UnLock(getServer)
	}
}

//Scheme return schema
func (s *ServiceDiscovery) Scheme() string {
	return schema
}

//Close 关闭
func (s *ServiceDiscovery) Close() {
	fmt.Println("Close")
	s.cli.Close()
}

//watcher 监听前缀
func (s *ServiceDiscovery) watcher() {
	var name string
	if s.prefixServerName != "" {
		name = s.prefixServerName
	} else {
		name = s.groupName
	}
	rch := s.cli.Watch(name)
	log.Printf("watching prefix:%s now...", name)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case DHT.PUT: //新增或修改
				fmt.Println("客户端在etcd中新增或修改 key：", string(ev.Kv.Key), " 对应的值为：", string(ev.Kv.Value))
				s.SetServiceList(string(ev.Kv.Key), string(ev.Kv.Value))
			case DHT.DELETE: //删除
				fmt.Printf("%s 节点下线\n", string(ev.Kv.Value))
				s.DelServiceList(string(ev.Kv.Value))
			}
		}
	}
}

//SetServiceList 设置服务地址
func (s *ServiceDiscovery) SetServiceList(key, val string) {
	//获取服务地址
	addr := resolver.Address{Addr: val}

	s.serverList.Store(val, addr)
	s.cc.UpdateState(resolver.State{Addresses: s.getServices()})
	log.Println(key, "模块添加服务，服务地址是：", val)
}

//DelServiceList 删除服务地址
func (s *ServiceDiscovery) DelServiceList(key string) {
	s.serverList.Delete(key)
	s.cc.UpdateState(resolver.State{Addresses: s.getServices()})
	log.Println("del key:", key)
}

//GetServices 获取服务地址
func (s *ServiceDiscovery) getServices() []resolver.Address {
	addrs := make([]resolver.Address, 0, 10)
	s.serverList.Range(func(k, v interface{}) bool {
		addrs = append(addrs, v.(resolver.Address))
		return true
	})
	return addrs
}

func (s *ServiceDiscovery) Listen() error {
	return s.cli.Listen()
}

func (s *ServiceDiscovery) GetGroup(groupName string) (DHT.RangeResponse, error) {
	return s.cli.GetGroup(groupName)
}
