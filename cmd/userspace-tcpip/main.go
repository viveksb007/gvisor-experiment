package main

import (
	"errors"
	"io"
	"log"
	"net"
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
	log.Println("Starting TCP server")
	tcpListener, e := gonet.ListenTCP(s, tcpip.FullAddress{
		Addr: serverAddress,
		Port: 8080,
	}, ipv4.ProtocolNumber)
	if e != nil {
		log.Fatalf("err in creating TCP listener = %v", e)
	}

	go func() {
		for {
			c, err := tcpListener.Accept()
			if err != nil {
				log.Fatalf("err in tcpListener Accept() = %v", err)
			}
			go handleConnection(c)
		}
	}()
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			log.Printf("Read error: %v", err)
			return
		}
		log.Printf("Remote Addr %s, Local Address %s\n", conn.RemoteAddr(), conn.LocalAddr())
		log.Printf("Received message on Server: %s\n", string(buf[:n]))

		_, err = conn.Write([]byte("Hey TCP Client"))
		if err != nil {
			log.Printf("Write error: %v", err)
			return
		}
	}
}

func startClient(s *stack.Stack, serverAddress tcpip.Address) {
	remoteAddress := tcpip.FullAddress{Addr: serverAddress, Port: 8080}

	testConn, err := connect(s, remoteAddress)
	if err != nil {
		log.Fatal("Unable to connect: ", err)
	}

	conn := gonet.NewTCPConn(testConn.wq, testConn.ep)
	defer conn.Close()

	message := "Hello, TCP Server!"
	_, err1 := conn.Write([]byte(message))
	if err1 != nil {
		log.Fatalf("Failed to write to connection: %v", err)
	}

	buf := make([]byte, 1024)
	n, err1 := conn.Read(buf)
	if err1 != nil {
		log.Fatalf("Failed to read from connection: %v", err)
	}

	log.Printf("Received response: %s\n", string(buf[:n]))
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
