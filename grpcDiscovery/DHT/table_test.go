package DHT

import (
	"net"
	"testing"
	"time"
)

func TestJsonNode(t *testing.T) {
	d, err := newDhtServer(8001, "127.0.0.1", "client", "127.0.0.1:3131")
	if err != nil {
		println(err.Error())
		return
	}
	go d.listen()
	n := node{
		id: "12123",
		addr: &net.UDPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 8000,
			Zone: "",
		},
		ServerType: "play",
		status:     true,
		grpcServer: "127.0.0.1:3131",
	}
	println(d.pingNode(&n))
	time.Sleep(time.Second * 10)
}
