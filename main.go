package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/lainio/err2"
	_ "github.com/lainio/err2/assert"
	"github.com/lainio/err2/try"
	greet "github.com/lainio/nats-first/grpc/greet/v1"
	"github.com/lainio/nats-first/nconn"
	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/encoders/protobuf"
)

var (
	name      = flag.String("name", "ville", "name for greet. protobuf message")
	json      = flag.Bool("json", false, "use JSON_ENCODER instead of PROTOBUF_ENCODER")
	publisher = flag.Bool("pub", false, "is publisher or if not THEN subscriber")
	stop      = flag.Bool("stop", false, "tells if we need to send stop at the end")
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

	codec := protobuf.PROTOBUF_ENCODER
	if *json {
		codec = nats.JSON_ENCODER
	}
	ec := nconn.New(codec)
	defer ec.Close()

	type person struct {
		Name    string
		Address string
		Age     int
	}

	if *json {
		try.To(doJSON(ec.EncodedConn))
	} else {
		try.To(doProtobuf(ec.EncodedConn))
	}
	time.Sleep(200 * time.Millisecond)
}

func doProtobuf(ec *nats.EncodedConn) (err error) {
	defer err2.Handle(&err)

	if *publisher {
		return doProtobufAsPub(ec)
	}
	return doProtobufAsSub(ec)
}

func doProtobufAsPub(ec *nats.EncodedConn) (err error) {
	defer err2.Handle(&err)

	glog.Infoln("doProtobufAsPub")

	sendCh := make(chan *greet.Request)
	try.To(ec.BindSendChan("hello", sendCh))

	var i int
	for i = 0; i < 10; i++ {
		stop := *stop && i == 9
		//stop := false
		s := fmt.Sprintf("%s_%d", *name, i)
		fmt.Println("name:", s)
		me := &greet.Request{Name: s, Stop: stop}
		sendCh <- me
	}
	glog.Infoln("done pub, i =", i)

	return nil
}

func doProtobufAsSub(ec *nats.EncodedConn) (err error) {
	defer err2.Handle(&err)

	glog.Infoln("doProtobufAsSub")

	recvCh := make(chan *greet.Request)
	try.To1(ec.BindRecvChan("hello", recvCh))

	for {
		who := <-recvCh
		glog.Infoln(who)
		if who.Stop {
			break
		}
	}
	glog.Infoln("done sub")

	return nil
}

func doJSON(ec *nats.EncodedConn) (err error) {
	defer err2.Handle(&err)

	recvCh := make(chan *person)
	try.To1(ec.BindRecvChan("hello", recvCh))

	sendCh := make(chan *person)
	try.To(ec.BindSendChan("hello", sendCh))

	me := &person{Name: "derek", Age: 22, Address: "140 New Montgomery Street"}

	sendCh <- me

	who := <-recvCh

	glog.Infoln(who)
	return nil
}
