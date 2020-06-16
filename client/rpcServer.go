package main

import (
	"client/rpcServer"
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"path"
	"strings"
)

const (
	SCRIPT_NAME      = "start.sh"
	SCRIPT_CONTENT   = "#!/bin/bash"
	DEFAULT_RPC_PORT = "19876"
)

type RpcServer struct{}

func (s *RpcServer) CreateScript(ctx context.Context, request *rpcServer.CreateScriptRequest) (response *rpcServer.CreateScriptResponse, err error) {
	response = new(rpcServer.CreateScriptResponse)
	if err = createContainerEntryPointScriptHandler(request); err != nil {
		response.ErrMessage = err.Error()
	}
	return
}

func createContainerEntryPointScriptHandler(req *rpcServer.CreateScriptRequest) (err error) {
	var (
		f               *os.File
		script          = path.Join(req.Pwd, SCRIPT_NAME)
		removeScriptErr error
		content         = make([]string, 0)
	)

	if _, err = os.Stat(script); err != nil {
		if os.IsExist(err) {
			removeScriptErr = os.Remove(script)
		}
	} else {
		removeScriptErr = os.Remove(script)
	}
	if removeScriptErr != nil {
		return errors.New(fmt.Sprintf("remove contianer entry point script error: %v", err))
	}

	if f, err = os.OpenFile(script, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755); err != nil {
		return errors.New(fmt.Sprintf("create contianer entry point script error: %v", err))
	}
	defer f.Close()

	content = append(append(content, SCRIPT_CONTENT), req.Cmd...)
	contentStr := strings.Join(content, "\n")

	if _, err = f.Write([]byte(contentStr)); err != nil {
		return errors.New(fmt.Sprintf("write to contianer entry point script error: %v", err))
	}

	return nil
}

func RunRpcServer() {
	lis, err := net.Listen("tcp", ":" + Conf.RpcPort)
	if err != nil {
		fmt.Println("rpcServer listen to port 19876 error:", err)
		return
	}
	s := grpc.NewServer()
	rpcServer.RegisterCreateContainerEntryPointScriptServer(s, &RpcServer{})
	reflection.Register(s)

	err = s.Serve(lis)
	if err != nil {
		fmt.Println("rpcServer start error:", err)
		return
	}
	lis.Accept()
}
