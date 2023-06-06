package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/smallnest/rpcx/client"
)

func main() {
	startTime := time.Now()

	type Args struct {
		FF   [][]byte
		Conf *pdfcpu.Configuration
	}

	type Reply struct {
		W []byte
	}

	var (
		addr = flag.String("addr", "localhost:8972", "server address")
		// etcdAddr = flag.String("etcdAddr", "192.168.21.71:2379", "etcd address")
		// basePath = flag.String("base", "/rpcx_pdfwtk", "prefix path")
	)
	// http.HandleFunc("/", index)
	// http.HandleFunc("/pdf/merge", pdfmerge)
	// http.HandleFunc("/pdf/merge/", pdfmerge)

	// serv := "0.0.0.0:8384"
	// if err := http.ListenAndServe(serv, nil); err != nil {
	// 	log.Fatal("ListenAndServe: ", err)
	// }
	flag.Parse()

	d, _ := client.NewPeer2PeerDiscovery("tcp@"+*addr, "")
	// d, _ := etcd_client.NewEtcdDiscovery(*basePath, "PDF", []string{*etcdAddr}, false, nil)
	xclient := client.NewXClient("PDF", client.Failtry, client.RandomSelect, d, client.DefaultOption)
	defer xclient.Close()

	inFiles := []string{"1.pdf", "2.pdf", "3.pdf"}
	outFile := "merge.pdf"
	var ff [][]byte
	for _, f := range inFiles {
		// fmt.Println(f)
		fb, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatal(err)
		}
		ff = append(ff, fb)
	}

	w, _ := os.Create(outFile)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	defer func() {
		// if err = w.Close(); err != nil {
		// 	return
		// }
		w.Close()
	}()

	conf := pdfcpu.NewDefaultConfiguration()
	conf.ValidationMode = pdfcpu.ValidationNone
	// conf.OwnerPW = "123456"
	// conf.UserPW = ""
	// conf.EncryptUsingAES = true
	// conf.EncryptKeyLength = 256
	// conf.Permissions = 204

	// rs := make([]io.ReadSeeker, len(ff))
	// for i, f := range ff {
	// 	rs[i] = f
	// }

	args := Args{
		FF:   ff,
		Conf: conf,
	}

	reply := &Reply{}

	// err := xclient.Call(context.Background(), "Merge", args, reply)
	call, err := xclient.Go(context.Background(), "Merge", args, reply, nil)
	if err != nil {
		log.Fatalf("failed to call: %v", err)
	}
	replyCall := <-call.Done
	if replyCall.Error != nil {
		log.Fatalf("failed to call: %v", replyCall.Error)
	} else {
		w.Write(reply.W)
		// fmt.Println(reply.W)
	}
	log.Print("duration:", time.Since(startTime))
}
