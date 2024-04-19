package DHT

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type messageType uint8
type requestType uint8

const (
	udpRequest   messageType = 'r' // 对端发送的信息
	udpResponse  messageType = 'b' // 对端返回的信息
	udpBroadcast messageType = 'g' // 广播信息
	register     requestType = 'r'
	offline      requestType = 'o'
	ping         requestType = 'p'
	getTable     requestType = 'a'
	pingTimeout              = time.Second * 10
)

type dhtNetWorker struct {
	selfNode        *node
	table           *dhtTable
	packetConn      net.PacketConn
	getCallbackData map[string]func(message)
	watchMap        map[string]chan<- node
	watch           chan node
	pingList        sync.Map //服务列表
}

type message struct {
	Y      messageType
	T      [10]byte
	Header requestHeader `json:"Header"`
	Body   []byte        `json:"Body"`
}
type requestHeader struct {
	BroadcastsNumber int         `json:"broadcastsNumber"`
	MsgType          requestType `json:"msgType"`
	RequestNodeId    string      `json:"requestNodeId"`
}
type registerBody struct {
	NodeId      string
	IpHost      string
	ServerType  string
	ServerName  string
	GrpcAddress string
}
type offlineBody struct {
	NodeId     string
	ServerType string
}
type tableBody struct {
	Nodes []nodeBody
}
type nodeBody struct {
	IpHost      string
	ServerType  string
	ServerName  string
	NodeId      string
	GrpcAddress string
}

func (dn *dhtNetWorker) listen() error {
	var cache [0x10000]byte
	for {
		n, addr, err := dn.packetConn.ReadFrom(cache[:])
		if err != nil {
			if ignoreReadFromError(err) {
				fmt.Println("对端已关闭")
				continue
			}
			return err
		}
		if n == len(cache) {
			fmt.Println("数据未完整读取")
			continue
		}
		dn.receiveMessage(cache[:n], addr)
	}
}
func (dn *dhtNetWorker) receiveMessage(data []byte, address net.Addr) {
	r := &message{}
	err := json.Unmarshal(data, r)
	if err != nil {
		fmt.Printf("message序列化错误:%s\n消息内容：%s,长度：%d;\n发送端地址为：%s\n", err.Error(), string(data), len(data), address.String())

		return
	}
	addr, ok := address.(*net.UDPAddr)
	if !ok {
		fmt.Println("address不是UDPAddr对象")
		return
	}
	switch r.Y {
	case udpBroadcast:
		go dn.queryMessage(r, addr)
	case udpRequest:
		go dn.queryMessage(r, addr)
	case udpResponse:
		v, ok := dn.getCallbackData[string(r.T[:])]
		if !ok {
			return
		}
		go v(*r)
		delete(dn.getCallbackData, string(r.T[:]))
	}
}
func (dn *dhtNetWorker) send(m []byte, addr net.Addr) error {
	n, err := dn.packetConn.WriteTo(m, addr)
	if n != len(m) {
		fmt.Printf("写入长度与数据长度不符，n:%d,m:%d", n, len(m))
		err = io.ErrShortWrite
		return err
	}
	if err != nil {
		return err
	}
	return nil
}
func (dn *dhtNetWorker) sendResponse(r message, addr net.Addr) error {
	r.Y = udpResponse
	m, err := json.Marshal(r)
	if err != nil {
		return err
	}

	return dn.send(m, addr)
}
func (dn *dhtNetWorker) sendRequest(ctx context.Context, r message, addr net.Addr) ([]byte, error) {
	responseChan := make(chan message, 1)
	r.Header.RequestNodeId = dn.selfNode.id
	dn.getCallbackData[string(r.T[:])] = func(m message) {
		responseChan <- m
	}
	var (
		result message
		err    error
	)
	m, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	err = dn.send(m, addr)
	if err != nil {
		return nil, err
	}
	select {
	case result = <-responseChan:
	case <-ctx.Done():
		err = errors.New("timeout")
	}
	return result.Body, err
}
func (dn *dhtNetWorker) close() {
	dn.packetConn.Close()
}

// 发送ping请求
func (dn *dhtNetWorker) pingNode(n *node) bool {
	m := message{
		Y: udpRequest,
		T: randT(),
		Header: requestHeader{
			MsgType: ping,
		},
		Body: nil,
	}
	con, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	_, err := dn.sendRequest(con, m, n.getAddr())
	if err != nil {
		return false
	}
	return true
}

// 处理ping消息
func (dn *dhtNetWorker) pingRequest(m *message, addr net.Addr) {
	dn.sendResponse(*m, addr)
}

// 对所有节点广播内容
func (dn *dhtNetWorker) sendBroadcastMessage(r *message, exceptionNodeIds ...string) {
	r.Y = udpBroadcast

	exceptionNodeIds = append(exceptionNodeIds, dn.selfNode.id)
	sort.Slice(exceptionNodeIds, func(i, j int) bool { return exceptionNodeIds[i] < exceptionNodeIds[j] })

	for _, n := range dn.table.nodeTable {
		if !n.status {
			continue
		}
		if inList(exceptionNodeIds, n.id) {
			continue
		}
		e := dn.sendResponse(*r, n.getAddr())
		if e != nil {
			println("广播出现异常：", e.Error())
		}
	}
}

// 按请求类型分配函数处理
func (dn *dhtNetWorker) queryMessage(r *message, address *net.UDPAddr) {
	switch r.Header.MsgType {
	case register: //上线
		dn.nodeUpload(r, address)
	case offline: // 下线
		dn.nodeOffLine(r)
	case ping:
		dn.pingRequest(r, address)
	case getTable:
		dn.sendTable(r, address)
	}
}

// 收到节点上线 广播消息
func (dn *dhtNetWorker) nodeUpload(r *message, address *net.UDPAddr) {
	RegisterBody := &registerBody{}
	err := json.Unmarshal(r.Body, RegisterBody)
	if err != nil {
		return
	}
	nodeInfo := dn.table.getNode(RegisterBody.NodeId)
	// 已存在的节点，不在处理
	if nodeInfo != nil {
		return
	}
	var nodeAddress *net.UDPAddr
	// 对方节点自己发送消息,这里需要返回自己的信息
	if r.Y == udpRequest {
		nodeAddress = address
		rg := message{Y: udpResponse, T: r.T}
		rg.Body, err = json.Marshal(nodeBody{
			IpHost:      dn.selfNode.addr.String(),
			ServerType:  dn.selfNode.ServerType,
			ServerName:  dn.selfNode.ServerName,
			NodeId:      dn.selfNode.id,
			GrpcAddress: dn.selfNode.grpcServer,
		})
		if err != nil {
			return
		}
		dn.sendResponse(rg, nodeAddress)
		fmt.Println("上线节点请求")
	} else if r.Y == udpBroadcast {
		// 其他节点广播的上线消息
		fmt.Println("节点上线 广播", RegisterBody.IpHost)
		nodeAddress, err = net.ResolveUDPAddr("udp", RegisterBody.IpHost)
		if err != nil {
			// 错误的udp地址
			println("错误的udp地址", err.Error())
			return
		}
	} else {
		// 异常情况
		fmt.Println("异常情况")
		return
	}
	nodeInfo = &node{
		addr:       nodeAddress,
		id:         RegisterBody.NodeId,
		ServerType: RegisterBody.ServerType,
		ServerName: RegisterBody.ServerName,
		grpcServer: RegisterBody.GrpcAddress,
		status:     true,
	}
	dn.table.append(nodeInfo)
	//fmt.Printf("%s模块节点上线，ip为%s, 节点id为 %d\n", nodeInfo.ServerType, nodeInfo.netAddr(), nodeInfo.id)
	//println("判断转发上线信息")
	if r.Header.BroadcastsNumber > 0 {
		// 广播次数上限
		r.Header.BroadcastsNumber--
		RegisterBody.IpHost = nodeAddress.String()
		r.Body, err = json.Marshal(RegisterBody)
		if err != nil {
			println("json失败")
			return
		}
		println("转发广播")
		dn.sendBroadcastMessage(r, RegisterBody.NodeId)
	}

	fmt.Println("发送上线消息")
	if dn.watch != nil {
		dn.watch <- *nodeInfo
	}
}

// 发送节点上线请求
func (dn *dhtNetWorker) register(targetServer net.Addr) (*node, error) {
	m := message{
		Y: udpRequest,
		T: randT(),
		Header: requestHeader{
			BroadcastsNumber: 2,
			MsgType:          register,
		},
		Body: nil,
	}

	r := registerBody{
		NodeId:      dn.selfNode.id,
		ServerType:  dn.selfNode.ServerType,
		ServerName:  dn.selfNode.ServerName,
		GrpcAddress: dn.selfNode.grpcServer,
	}
	m.Body, _ = json.Marshal(r)
	con, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	data, err := dn.sendRequest(con, m, targetServer)
	if err != nil {
		fmt.Printf("发起请求错误%s\n", err.Error())
		return nil, err
	}
	n := &nodeBody{}
	err = json.Unmarshal(data, n)
	if err != nil {
		fmt.Printf("序列化错误%s\n", err.Error())
		return nil, err
	}
	return &node{
		id:         n.NodeId,
		addr:       targetServer.(*net.UDPAddr),
		ServerType: n.ServerType,
		ServerName: n.ServerName,
		status:     true,
		grpcServer: n.GrpcAddress,
	}, nil
}

// 收到节点下线 广播消息
func (dn *dhtNetWorker) nodeOffLine(r *message) {
	OfflineBody := &offlineBody{}
	err := json.Unmarshal(r.Body, OfflineBody)
	if err != nil {
		return
	}
	nodeInfo := dn.table.getNode(OfflineBody.NodeId)
	if nodeInfo == nil {
		// 这个服务就没在此存在，或者已经下线了,属于无效消息
		return
	}
	//fmt.Printf("%s节点下线,节点id：%d\n", nodeInfo.addr.String(), nodeInfo.id)
	if nodeInfo.status && dn.pingNode(nodeInfo) {
		// 可以ping通，服务状态良好，属于无效消息
		return
	} else {
		nodeInfo.status = false
		r.Header.BroadcastsNumber = 1
	}
	if r.Header.BroadcastsNumber == 0 {
		// 广播次数上限
		return
	}
	r.Header.BroadcastsNumber--
	dn.sendBroadcastMessage(r, OfflineBody.NodeId)

}

// 发送节点下线 广播消息
func (dn *dhtNetWorker) offline(serverType string, nodeId string) {
	m := message{
		Y: udpBroadcast,
		T: randT(),
		Header: requestHeader{
			BroadcastsNumber: 2,
			MsgType:          offline,
		},
		Body: nil,
	}
	o := offlineBody{
		NodeId:     nodeId,
		ServerType: serverType,
	}
	m.Body, _ = json.Marshal(o)
	dn.table.delete(nodeId)
	dn.sendBroadcastMessage(&m)
}

// 获取别的节点table
func (dn *dhtNetWorker) getNodeTable(n net.Addr) {
	m := message{
		Y: udpRequest,
		T: randT(),
		Header: requestHeader{
			MsgType: getTable,
		},
		Body: nil,
	}
	con, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	data, err := dn.sendRequest(con, m, n)
	if err != nil {
		fmt.Println("发送请求错误", err.Error())
		return
	}
	t := &tableBody{}
	err = json.Unmarshal(data, t)
	if err != nil {
		fmt.Println("table序列化错误", err.Error())
		return
	}
	dn.table.initTable(*t)
}

// 发送本节点的table信息
func (dn *dhtNetWorker) sendTable(m *message, addr *net.UDPAddr) {
	var cache []nodeBody
	for _, n := range dn.table.nodeTable {
		cache = append(cache, nodeBody{
			IpHost:      n.addr.String(),
			ServerType:  n.ServerType,
			NodeId:      n.id,
			GrpcAddress: n.grpcServer,
		})
	}
	t := tableBody{
		Nodes: cache,
	}
	d, err := json.Marshal(t)
	if err != nil {
		fmt.Println("table序列化错误", err.Error())
		return
	}
	m.Body = d
	m.Y = udpResponse
	err = dn.sendResponse(*m, addr)
	if err != nil {
		fmt.Println("发送包失败:", err.Error())
	}
}

// 注册监听的类型
func (dn *dhtNetWorker) registerWatch(watchType string, result chan<- node) {
	dn.watchMap[watchType] = result
}

// 检查节点状态
func (dn *dhtNetWorker) checkStatus(nodes []*node) {
	for _, n := range nodes {
		if n.status && !dn.pingNode(n) {
			n.status = false
			dn.offline(n.ServerType, n.id)
			if dn.watch != nil {
				dn.watch <- *n
			}
		}
	}
}

// 通知服务的上下线
func (dn *dhtNetWorker) sendWatch(n node) {
	v, ok := dn.watchMap[n.ServerType]
	if !ok {
		return
	}
	v <- n

}

func randT() (result [10]byte) {
	for i := 0; i < 10; i++ {
		result[i] = byte(rand.Intn(254))
	}
	return
}

func newDhtServer(port int, serverIp, serverType, serverName, grpcAddress string) (*dhtNetWorker, error) {
	selfIp := net.JoinHostPort(serverIp, strconv.Itoa(port))
	nodeAddress, err := net.ResolveUDPAddr("udp", selfIp)
	if err != nil {
		// 错误的udp地址
		return nil, err
	}
	selfNode := node{
		id:         fmt.Sprintf("%s://%s:%d||%s", serverType, serverIp, port, grpcAddress),
		addr:       nodeAddress,
		ServerType: serverType,
		ServerName: serverName,
		status:     true,
		grpcServer: grpcAddress,
	}
	packetConn, err := net.ListenPacket("udp", selfIp)
	if err != nil {
		return nil, err
	}
	dht := dhtNetWorker{
		selfNode: &selfNode,
		table: &dhtTable{
			table:     make(map[string][]*node),
			nodeTable: make(map[string]*node),
		},
		packetConn:      packetConn,
		getCallbackData: make(map[string]func(message)),
		watch:           nil,
	}
	dht.table.append(&selfNode)
	return &dht, nil
}
