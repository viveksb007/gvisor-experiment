package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/loopback"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"
)

const NICID = 1

func main() {
	s := stack.New(stack.Options{
		NetworkProtocols:   []stack.NetworkProtocolFactory{ipv4.NewProtocol},
		TransportProtocols: []stack.TransportProtocolFactory{tcp.NewProtocol},
	})

	if err := s.CreateNIC(1, loopback.New()); err != nil {
		log.Fatalf("Failed to create NIC: %v", err)
	}

	protocolAddr := tcpip.ProtocolAddress{
		Protocol: ipv4.ProtocolNumber,
		AddressWithPrefix: tcpip.AddressWithPrefix{
			Address:   tcpip.AddrFromSlice(net.IPv4(192, 168, 1, 1).To4()),
			PrefixLen: 32,
		},
	}

	if err := s.AddProtocolAddress(NICID, protocolAddr, stack.AddressProperties{}); err != nil {
		log.Fatalf("Failed to add protocol address: %v", err)
	}

	s.SetRouteTable([]tcpip.Route{
		{
			NIC:         NICID,
			Destination: header.IPv4EmptySubnet,
		},
	})

	startServer(s, protocolAddr.AddressWithPrefix.Address)

	for i := 0; i < 3; i++ {
		startClient(s, protocolAddr.AddressWithPrefix.Address)
		time.Sleep(100 * time.Millisecond)
	}
}

func startServer(s *stack.Stack, serverAddress tcpip.Address) {
	tcpListener, e := gonet.ListenTCP(s, tcpip.FullAddress{
		Addr: serverAddress,
		Port: 8080,
	}, ipv4.ProtocolNumber)
	if e != nil {
		log.Fatalf("NewListener() = %v", e)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request from client %v\n", r.RemoteAddr)
		fmt.Fprintf(w, "yo man")
	})

	log.Println("Starting http server on port 8080")
	go func() {
		if err := http.Serve(tcpListener, nil); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
}

func startClient(s *stack.Stack, serverAddress tcpip.Address) {
	remoteAddress := tcpip.FullAddress{Addr: serverAddress, Port: 8080}

	testConn, err := connect(s, remoteAddress)
	if err != nil {
		log.Fatal("Unable to connect: ", err)
	}

	conn := gonet.NewTCPConn(testConn.wq, testConn.ep)
	defer conn.Close()

	log.Printf("TCP connection to %v is successful\n", conn.RemoteAddr().String())

	client := &http.Client{
		Transport: &customRoundTripper{conn: conn},
	}

	resp, err1 := client.Get(fmt.Sprintf("http://%s:8080", serverAddress.WithPrefix().Address.String()))
	if err1 != nil {
		log.Fatalf("Failed to send HTTP request: %v", err1)
	}
	defer resp.Body.Close()

	body, err1 := io.ReadAll(resp.Body)
	if err1 != nil {
		log.Fatalf("Failed to read HTTP response: %v", err1)
	}

	log.Printf("Received response: %s\n", string(body))
}

type testConnection struct {
	wq *waiter.Queue
	ep tcpip.Endpoint
}

func connect(s *stack.Stack, addr tcpip.FullAddress) (*testConnection, tcpip.Error) {
	wq := &waiter.Queue{}
	ep, err := s.NewEndpoint(tcp.ProtocolNumber, ipv4.ProtocolNumber, wq)
	if err != nil {
		return nil, err
	}

	entry, ch := waiter.NewChannelEntry(waiter.WritableEvents)
	wq.EventRegister(&entry)

	err = ep.Connect(addr)
	if _, ok := err.(*tcpip.ErrConnectStarted); ok {
		<-ch
		err = ep.LastError()
	}
	if err != nil {
		return nil, err
	}

	log.Println(ep.GetLocalAddress())
	log.Println(ep.GetRemoteAddress())

	return &testConnection{wq, ep}, nil
}

type customRoundTripper struct {
	conn net.Conn
}

func (rt *customRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	err := req.Write(rt.conn)
	if err != nil {
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(rt.conn), req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
