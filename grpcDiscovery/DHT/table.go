package DHT

import (
	"net"
)

var (
	localhost = net.ParseIP("127.0.0.1")
)

/*
todo 这里需要改成 红黑树存储节点，
	目的：为了有序广播，广播向自己的上下节点广播，排除掉消息来源方向。
*/
type dhtTable struct {
	table     map[string][]*node
	nodeTable map[string]*node // key 为 nodeId
}

func (dT *dhtTable) initTable(body tableBody) {
	for _, n := range body.Nodes {
		nodeAddress, err := net.ResolveUDPAddr("udp", n.IpHost)
		if err != nil {
			// 错误的udp地址
			println("initTable：错误的udp地址")
			continue
		}
		if nodeAddress.IP == nil {
			nodeAddress.IP = localhost
		}

		no := node{
			id:         n.NodeId,
			addr:       nodeAddress,
			ServerType: n.ServerType,
			status:     true,
			grpcServer: n.GrpcAddress,
		}
		dT.append(&no)
	}
}
func (dT *dhtTable) getNode(nodeId string) *node {
	value, ok := dT.nodeTable[nodeId]
	if !ok {
		return nil
	}
	if !value.status {
		return nil
	}
	return value
}
func (dT *dhtTable) get(ServerType string) []*node {
	value, ok := dT.table[ServerType]
	if !ok {
		return []*node{}
	}
	return value

}
func (dT *dhtTable) append(serverInfo *node) bool {
	ns, ok := dT.nodeTable[serverInfo.id]
	if ok {
		ns.grpcServer = serverInfo.grpcServer
		ns.status = true
		ns.ServerType = serverInfo.ServerType
		return true
	}
	_, ok = dT.table[serverInfo.ServerType]
	if !ok {
		dT.table[serverInfo.ServerType] = []*node{serverInfo}
		dT.nodeTable[serverInfo.id] = serverInfo
		return true
	}
	for _, n := range dT.table[serverInfo.ServerType] {
		if n.id == serverInfo.id {
			return false
		}
	}
	dT.table[serverInfo.ServerType] = append(dT.table[serverInfo.ServerType], serverInfo)
	dT.nodeTable[serverInfo.id] = serverInfo
	return true
}
func (dT *dhtTable) delete(nodeId string) bool {
	dT.nodeTable[nodeId].status = false
	//value := dT.get(ServerType)
	//index := -1
	//for i, n := range value {
	//	if n.id == nodeId {
	//		index = i
	//		break
	//	}
	//}
	//if index == -1 {
	//	return false
	//}
	//delete(dT.nodeTable, nodeId)
	//if index == len(value)-1 {
	//	dT.table[ServerType] = value[:len(value)-1]
	//	return true
	//}
	//dT.table[ServerType] = append(value[:index], value[index+1:]...)
	return true
}
func (dT *dhtTable) length() int {
	result := 0
	for _, n := range dT.nodeTable {
		if n.status {
			result++
		}
	}
	return result
}
func (dT *dhtTable) getMd5() []byte {
	return nil
}

type node struct {
	id         string
	addr       *net.UDPAddr
	ServerType string
	ServerName string
	status     bool // 在线状态
	grpcServer string
}

func (n node) getAddr() net.Addr {
	return n.addr
}
func (n *node) netAddr() []byte {
	return []byte(n.addr.String())
}
