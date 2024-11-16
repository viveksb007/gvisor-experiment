package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/samber/lo"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/fdbased"
	"gvisor.dev/gvisor/pkg/tcpip/link/rawfile"
	"gvisor.dev/gvisor/pkg/tcpip/link/sniffer"
	"gvisor.dev/gvisor/pkg/tcpip/link/tun"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"
)

func main() {
	clientAddr := tcpip.ProtocolAddress{
		Protocol: ipv4.ProtocolNumber,
		AddressWithPrefix: tcpip.AddressWithPrefix{
			Address:   tcpip.AddrFromSlice(net.IPv4(192, 168, 1, 2).To4()),
			PrefixLen: 24,
		},
	}

	var domainName string
	flag.StringVar(&domainName, "domain", "google.com", "Domain Name")
	flag.Parse()
	ips := lo.Must1(net.LookupIP(domainName))

	serverAddr := tcpip.FullAddress{
		NIC:  1,
		Addr: tcpip.AddrFromSlice(net.ParseIP(ips[0].String()).To4()),
	}
	serverAddr.Port = uint16(80)

	s := stack.New(stack.Options{
		NetworkProtocols:   []stack.NetworkProtocolFactory{ipv4.NewProtocol},
		TransportProtocols: []stack.TransportProtocolFactory{tcp.NewProtocol},
	})

	var tunName = "tun0"
	mtu := lo.Must1(rawfile.GetMTU(tunName))
	fd := lo.Must1(tun.Open(tunName))

	linkEP := lo.Must1(fdbased.New(&fdbased.Options{FDs: []int{fd}, MTU: mtu}))
	lo.Must0(s.CreateNIC(1, sniffer.New(linkEP)))
	lo.Must0(s.AddProtocolAddress(1, clientAddr, stack.AddressProperties{}))

	s.SetRouteTable([]tcpip.Route{
		{
			NIC:         1,
			Destination: header.IPv4EmptySubnet,
		},
	})

	testConn, err := connect(s, serverAddr)
	if err != nil {
		log.Fatal("Unable to connect: ", err)
	}

	conn := gonet.NewTCPConn(testConn.wq, testConn.ep)
	defer conn.Close()

	log.Printf("TCP connection to %v is successful\n", conn.RemoteAddr().String())

	client := &http.Client{
		Transport: &customRoundTripper{conn: conn},
	}

	resp := lo.Must1(client.Get(fmt.Sprintf("http://%s", domainName)))
	defer resp.Body.Close()

	body := lo.Must1(io.ReadAll(resp.Body))

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
