package nconn

import (
	"github.com/golang/glog"
	_ "github.com/lainio/err2/assert"
	"github.com/lainio/err2/try"
	nats "github.com/nats-io/nats.go"
)

type NConn struct {
	*nats.EncodedConn
}

func New(codec string) NConn {
	glog.V(4).Infoln("new nats conn")
	nc := try.To1(nats.Connect(nats.DefaultURL))
	return NConn{try.To1(nats.NewEncodedConn(nc, codec))}
}

func (nc NConn) Pub(s string)  {
	
}
