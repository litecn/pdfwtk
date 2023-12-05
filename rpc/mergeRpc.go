package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"pdfwtk/pkg"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/smallnest/rpcx/server"
)

var (
	// addr = flag.String("addr", "localhost:8972", "server address")
	addr     = flag.String("addr", GetIp()+":"+GetPort(), "server address")
	etcdAddr = flag.String("etcdAddr", EtcdClust(), "etcd address")
	basePath = flag.String("base", "/rpcx_pdfwtk", "prefix path")
)

type Args struct {
	FF   [][]byte
	Conf *model.Configuration
}

type Reply struct {
	W []byte
}

type PDF struct{}

func GetIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err)
		return ""
	}
	for _, address := range addrs {
		// check loopback
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func EtcdClust() string {
	etcds := []string{"192.168.21.71:23790", "192.168.21.72:23790", "192.168.21.73:23790", "192.168.21.71:2379", "192.168.21.72:2379", "192.168.21.73:2379"}
	for _, etcd := range etcds {
		if IsListened(etcd) {
			log.Println("etcd:", etcd)
			return etcd
		}
	}
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

func GetPort() string {
	startPort := 8972
	endPort := startPort + 10
	for i := startPort; i < endPort; i++ {
		if IsPortAvailable(i) {
			return strconv.Itoa(i)
		}
	}
	return ""
}

func IsPortAvailable(port int) bool {
	address := fmt.Sprintf("%s:%d", "0.0.0.0", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		// log.Printf("port %s is taken: %s", address, err)
		return false
	}
	defer listener.Close()
	return true
}

// the second parameter is not a pointer
func (t *PDF) Merge(ctx context.Context, args Args, reply *Reply) error {
	startTime := time.Now()
	// time.Sleep(time.Second * 3)
	conf := args.Conf
	infiles := []io.ReadSeeker{}

	if len(args.FF) > 0 {
		for _, f := range args.FF {
			infile := io.ReadSeeker(bytes.NewReader(f))
			infiles = append(infiles, infile)
		}
	}

	// log.Println("files:", len(infiles))
	w := bytes.NewBuffer(reply.W)
	ww := io.Writer(w)
	err := pkg.MergeRaw(infiles, ww, conf)
	if err != nil {
		log.Print(err)
	}

	reply.W = w.Bytes()

	if conf.OwnerPW != "" || conf.UserPW != "" {
		rs := bytes.NewReader(reply.W)
		w.Reset()
		err = api.Encrypt(rs, ww, conf)
		if err != nil {
			log.Print(err)
			// return err
		} else {
			reply.W = w.Bytes()
		}
	}

	log.Println("files:", len(infiles), " bytes:", humanize.Bytes(uint64(len(w.Bytes()))))
	// log.Println("Duration:", humanize.Time(startTime))
	log.Println("Duration:", time.Since(startTime).String())
	// println()
	return err
}

func main() {
	flag.Parse()

	s := server.NewServer()
	addRegistryPlugin(s)

	//s.Register(new(Arith), "")
	s.RegisterName("PDF", new(PDF), "")

	log.Println("Server:", *addr)
	err := s.Serve("tcp", *addr)
	if err != nil {
		panic(err)
	}

	defer s.UnregisterAll()
}

func addRegistryPlugin(s *server.Server) {
	r := &serverplugin.EtcdRegisterPlugin{
		ServiceAddress: "tcp@" + *addr,
		EtcdServers:    []string{*etcdAddr},
		BasePath:       *basePath,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Second * 3,
	}
	err := r.Start()
	if err != nil {
		log.Fatal(err)
	}
	s.Plugins.Add(r)
}
