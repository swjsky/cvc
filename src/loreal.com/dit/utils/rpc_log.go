package utils

import (
	"log"
	"loreal.com/dit/cmd/logger/modle"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"
)

// CallRPCLog call rpc log
func CallRPCLog(addr string, logger *modle.Logger) (reply string, err error) {
	err = callRPC("App.Send", addr, logger, &reply)
	return
}

func callRPC(method, addr string, args, reply interface{}) error {
	client, err := jsonRPCClient("tcp", addr, time.Second)
	if err != nil {
		log.Println("cannot connect to", addr, "err:", err.Error())
		return err
	}
	err = client.Call(method, args, reply)
	if err != nil {
		return err
	}
	return nil
}

func jsonRPCClient(network, address string, timeout time.Duration) (*rpc.Client, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	return jsonrpc.NewClient(conn), err
}
