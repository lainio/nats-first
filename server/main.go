package main

import (
	"context"
	"flag"
	"os"

	"github.com/findy-network/findy-common-go/rpc"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
	_ "github.com/lainio/err2/assert" // we want an --asserter flag
	"github.com/lainio/err2/try"
	in "github.com/lainio/nats-first/grpc/in/v1"
	"github.com/lainio/nats-first/nconn"
	"github.com/nats-io/nats.go/encoders/protobuf"
	"google.golang.org/grpc"
)

const (
	CMD_SUBJECT = "cmd"
)

var (
	user       = flag.String("user", "findy-root", "test user name")
	serverAddr = flag.String("addr", "localhost", "agency host gRPC address")
	port       = flag.Int("port", 50051, "agency host gRPC port")
	noTLS      = flag.Bool("no-tls", false, "do NOT use TLS and cert files (hard coded)")
)

func main() {
	os.Args = append(os.Args,
		"-logtostderr",
	)
	glog.CopyStandardLogTo("ERROR") // for err2 binging

	defer err2.Catch(err2.ToStderr)

	flag.Parse()

	var pki *rpc.PKI
	if !*noTLS {
		pki = rpc.LoadPKI("./cert")
		glog.V(3).Infof("starting gRPC server with\ncrt:\t%s\nkey:\t%s\nclient:\t%s",
			pki.Server.CertFile, pki.Server.KeyFile, pki.Client.CertFile)
	}

	ec := nconn.New(protobuf.PROTOBUF_ENCODER)
	sendCh := make(chan *in.Cmd, 1)
	try.To(ec.BindSendChan(CMD_SUBJECT, sendCh))

	recvCh := make(chan *in.Cmd, 1)
	try.To1(ec.BindRecvChan(CMD_SUBJECT, recvCh))

	rpc.Serve(&rpc.ServerCfg{
		NoAuthorization: *noTLS,

		Port: *port,
		PKI:  pki,
		Register: func(s *grpc.Server) error {
			in.RegisterInServiceServer(s, &inService{
				NConn:  ec,
				Root:   *user,
				inCmd:  recvCh,
				outCmd: sendCh,
			})
			glog.V(10).Infoln("GRPC registration all done")
			return nil
		},
	})
}

type inService struct {
	in.UnimplementedInServiceServer

	nconn.NConn

	Root   string
	outCmd chan *in.Cmd
	inCmd  chan *in.Cmd
}

// ListenCmd implements v1.InServiceServer.
func (d inService) ListenCmd(icmd *in.Cmd, server in.InService_ListenCmdServer) (err error) {
	defer err2.Handle(&err)

	glog.V(3).Infoln("start listen")

	ctx := server.Context()
	recvCh := make(<-chan *in.Cmd, 1)
	try.To1(d.BindRecvChan(CMD_SUBJECT, recvCh))

loop:
	for {
		select {
		case <-ctx.Done():
			glog.V(1).Infoln("done")
			break loop
		case cmd := <-recvCh:
			glog.Infoln(cmd)
			try.To(server.Send(&in.CmdStatus{CmdType: icmd.Type}))
		}
	}
	return nil
}

func (d inService) EnterCmd(_ context.Context, received *in.Cmd) (_ *in.CmdStatus, err error) {
	defer err2.Handle(&err)

	glog.V(1).Info("--- enter Enter()", received.CmdID, received.Text)
	assert.CNotNil(d.outCmd)

	d.outCmd <- received

	return &in.CmdStatus{
		CmdID:   received.CmdID,
		Type:    in.CmdStatus_STATUS,
		CmdType: received.Type,
		Info: &in.CmdStatus_Ok{
			Ok: &in.CmdStatus_OKResult{
				Data: received.Text,
			},
		},
	}, nil
}
