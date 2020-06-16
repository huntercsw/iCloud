package rpcServer

import (
	"context"
	"github.com/docker/docker/api/types/strslice"
	"google.golang.org/grpc"
	"iCloud/log"
)

func CreateEntryPointScript(ip string, port string, pwd string, cmds strslice.StrSlice) (err error) {
	var (
		conn *grpc.ClientConn
		clientRpc = ip + ":" + port
		m = "rpcServer.entryPointScriptCreate()"
		rsp = new(CreateScriptResponse)
	)

	if conn, err = grpc.Dial(clientRpc, grpc.WithInsecure()); err != nil {
		log.Logger.Errorf("%s error, connect to client %s error: %v", m, clientRpc, err)
		return
	}
	defer conn.Close()

	c := NewCreateContainerEntryPointScriptClient(conn)
	if rsp, err = c.CreateScript(context.TODO(), &CreateScriptRequest{
		Pwd:pwd,
		Cmd:cmds,
	}); err != nil {
		log.Logger.Errorf("%s error, call client %s to create entry point script by grpc error: %v", m, clientRpc, err)
		log.Logger.Errorf("%s error: %v", m, rsp.ErrMessage)
		return
	}

	return nil
}
