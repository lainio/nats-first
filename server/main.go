package main

import (
	"context"
	"flag"
	"os"

	"github.com/findy-network/findy-common-go/rpc"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	_ "github.com/lainio/err2/assert" // we want an --asserter flag
	in "github.com/lainio/nats-first/grpc/in/v1"
	"google.golang.org/grpc"
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
	rpc.Serve(&rpc.ServerCfg{
		NoAuthorization: *noTLS,

		Port: *port,
		PKI:  pki,
		Register: func(s *grpc.Server) error {
			in.RegisterInServiceServer(s, &inService{Root: *user})
			glog.V(10).Infoln("GRPC registration all done")
			return nil
		},
	})
}

type inService struct {
	in.UnimplementedInServiceServer
	Root string
}

// ListenCmd implements v1.InServiceServer.
func (d inService) ListenCmd(*in.Cmd, in.InService_ListenCmdServer) (err error) {
	defer err2.Handle(&err)
	return nil
}

func (d inService) EnterCmd(context.Context, *in.Cmd) (_ *in.CmdStatus, err error) {
	defer err2.Handle(&err)
	glog.V(1).Info("enter Enter()")
	return &in.CmdStatus{
		CmdID:   0,
		Type:    0,
		CmdType: 0,
		Info: &in.CmdStatus_Ok{
			Ok: &in.CmdStatus_OKResult{
				Data: "what",
			},
		},
	}, nil
}

/*
func (d inService) Enter(ctx context.Context, cmd *in.Cmd) (cr *in.CmdReturn, err error) {
	defer err2.Handle(&err)

	glog.V(1).Info("enter Enter()")
	if !*noTLS {
		user := jwt.User(ctx)

		if user != d.Root {
			return &in.CmdReturn{Type: cmd.Type}, errors.New("access right")
		}
	}

	glog.V(3).Infoln("dev ops cmd", cmd.Type)
	cmdReturn := &in.CmdReturn{Type: cmd.Type}

	switch cmd.Type {
	case in.Cmd_PING:
		response := fmt.Sprintf("%s, ping ok", "TEST")
		cmdReturn.Response = &in.CmdReturn_Ping{Ping: response}
	case in.Cmd_LOGGING:
		//agencyCmd.ParseLoggingArgs(cmd.GetLogging())
		//response = fmt.Sprintf("logging = %s", cmd.GetLogging())
	case in.Cmd_COUNT:
		response := fmt.Sprintf("%d/%d cloud agents",
			100, 1000)
		cmdReturn.Response = &in.CmdReturn_Ping{Ping: response}
	}
	return cmdReturn, nil
}
*/
