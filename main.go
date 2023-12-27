package main

import (
	"flag"
	"os"

	"github.com/golang/glog"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	pb "github.com/nats-first/grpc-gen/main.pb.go"
	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/encoders/protobuf"
)

var (
	json = flag.Bool("json", false, "use JSON_ENCODER instead of PROTOBUF_ENCODER")
)

// EncodedConn can Publish any raw Go type using the registered Encoder
type person struct {
	Name    string
	Address string
	Age     int
}

func main() {
	glog.CopyStandardLogTo("ERROR") // for err2
	os.Args = append(os.Args,
		"-logtostderr",
		"-v=1",
	)

	defer err2.Catch()

	flag.Parse()

	nc := try.To1(nats.Connect(nats.DefaultURL))
	codec := protobuf.PROTOBUF_ENCODER
	if *json {
		codec = nats.JSON_ENCODER
	}
	ec := try.To1(nats.NewEncodedConn(nc, codec))
	defer ec.Close()

	type person struct {
		Name    string
		Address string
		Age     int
	}

	if *json {
		doJson(ec)
	} else {
		doProtobuf(ec)
	}
}

func doProtobuf(ec *nats.EncodedConn) {
	recvCh := make(chan *person)
	ec.BindRecvChan("hello", recvCh)

	sendCh := make(chan *person)
	ec.BindSendChan("hello", sendCh)

	me := &pb.GreetRequest{Name:"tero"}

	sendCh <- me

	who := <-recvCh

	glog.Infoln(who)
}

func doJson(ec *nats.EncodedConn) {
	recvCh := make(chan *person)
	ec.BindRecvChan("hello", recvCh)

	sendCh := make(chan *person)
	ec.BindSendChan("hello", sendCh)

	me := &person{Name: "derek", Age: 22, Address: "140 New Montgomery Street"}

	sendCh <- me

	who := <-recvCh

	glog.Infoln(who)
}
