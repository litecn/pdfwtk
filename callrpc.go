package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	etcd_client "github.com/rpcxio/rpcx-etcd/client"
	"github.com/smallnest/rpcx/client"
)

type Args struct {
	FF   [][]byte
	Conf *pdfcpu.Configuration
}

type Reply struct {
	W []byte
}

var (
	etcdAddr = flag.String("etcdAddr", EtcdClust(), "etcd address")
	basePath = flag.String("base", "/rpcx_pdfwtk", "prefix path")
)

func EtcdClust() string {
	etcds := []string{"192.168.21.71:23790", "192.168.21.72:23790", "192.168.21.73:23790", "192.168.21.71:2379", "192.168.21.72:2379", "192.168.21.73:2379"}
	for _, etcd := range etcds {
		if IsListened(etcd) {
			// log.Println("etcd:", etcd)
			return etcd
		}
	}
	log.Println("not found etcd!")
	return ""
}

func IsListened(address string) bool {
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return false
	} else {
		if conn != nil {
			_ = conn.Close()
			return true
		} else {
			return false
		}
	}
}

func CallRpc(inFiles []string, outFile string, conf *pdfcpu.Configuration) (*Reply, error) {
	flag.Parse()
	d, _ := etcd_client.NewEtcdDiscovery(*basePath, "PDF", []string{*etcdAddr}, false, nil)
	xclient := client.NewXClient("PDF", client.Failtry, client.RandomSelect, d, client.DefaultOption)
	defer xclient.Close()
	var ff [][]byte
	for _, f := range inFiles {
		// fmt.Println(f)
		fb, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		ff = append(ff, fb)
	}

	args := Args{
		FF:   ff,
		Conf: conf,
	}

	reply := &Reply{}

	// err := xclient.Call(context.Background(), "Merge", args, reply)
	call, err := xclient.Go(context.Background(), "Merge", args, reply, nil)
	if err != nil {
		log.Fatalf("failed to call: %v", err)
		return nil, err
	}
	replyCall := <-call.Done
	if replyCall.Error != nil {
		log.Fatalf("failed to call: %v", replyCall.Error)
		return nil, err
	} else {
		return reply, nil
		// fmt.Println(reply.W)
	}
}
