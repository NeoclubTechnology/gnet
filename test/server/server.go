package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"gitlab.neoclub.cn/NeoGo/gnet"
	"gitlab.neoclub.cn/NeoGo/gnet/pool/goroutine"

	"gitlab.neoclub.cn/NeoGo/gnet/test/protocol"
)

type customCodecServer struct {
	*gnet.EventServer
	addr       string
	multicore  bool
	async      bool
	codec      gnet.NICodec
	workerPool *goroutine.Pool
}

func (cs *customCodecServer) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	log.Printf("Test codec server is listening on %s (multi-cores: %t, loops: %d)\n",
		srv.Addr.String(), srv.Multicore, srv.NumEventLoop)
	return
}

func (cs *customCodecServer) React(frame interface{}, c gnet.Conn) (out interface{}, action gnet.Action) {
	data := frame.(*protocol.NimProtocol)
	fmt.Println("frame:", string(data.Body), data.CmdId)

	// store customize protocol header param using `c.SetContext()`
	//item := protocol.CustomLengthFieldProtocol{Version: protocol.DefaultProtocolVersion, ActionType: protocol.ActionData}
	//c.SetContext(item)

	if cs.async {
		//data := append([]byte{}, frame...)
		_ = cs.workerPool.Submit(func() {
			c.AsyncWrite(frame)
		})
		return
	}
	out = frame
	return
}

func testCustomCodecServe(addr string, multicore, async bool, codec gnet.NICodec) {
	var err error
	codec = &protocol.NimProtocol{}
	cs := &customCodecServer{addr: addr, multicore: multicore, async: async, codec: codec, workerPool: goroutine.Default()}
	objPool := &sync.Pool{
		New: func() interface{} {
			return &protocol.NimProtocol{}
		},
	}
	err = gnet.Serve(cs, addr, gnet.WithMulticore(multicore), gnet.WithTCPKeepAlive(time.Minute*5),
		gnet.WithNCodec(codec), gnet.WithObjPool(objPool))
	if err != nil {
		panic(err)
	}
}

func main() {
	var port int
	var multicore bool

	// Example command: go run server.go --port 9000 --multicore=true
	flag.IntVar(&port, "port", 9000, "server port")
	flag.BoolVar(&multicore, "multicore", true, "multicore")
	flag.Parse()
	addr := fmt.Sprintf("tcp://:%d", port)

	testCustomCodecServe(addr, multicore, false, nil)
}
