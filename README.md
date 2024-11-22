## gvisor experiments

1. Userspace TCP client and server (https://viveksb007.github.io/2024/10/gvisor-userspace-tcp-server-client/)

    `go run cmd/userspace-tcpip/main.go`
    ```
    2024/10/19 19:19:01 Starting TCP server
    2024/10/19 19:19:01 {0 192.168.1.1 46896 } <nil>
    2024/10/19 19:19:01 {0 192.168.1.1 8080 } <nil>
    2024/10/19 19:19:01 Remote Addr 192.168.1.1:46896, Local Address 192.168.1.1:8080
    2024/10/19 19:19:01 Received message on Server: Hello, TCP Server!
    2024/10/19 19:19:01 Received response: Hey TCP Client
    2024/10/19 19:19:02 {0 192.168.1.1 46897 } <nil>
    2024/10/19 19:19:02 {0 192.168.1.1 8080 } <nil>
    2024/10/19 19:19:02 Remote Addr 192.168.1.1:46897, Local Address 192.168.1.1:8080
    2024/10/19 19:19:02 Received message on Server: Hello, TCP Server!
    2024/10/19 19:19:02 Received response: Hey TCP Client
    2024/10/19 19:19:02 {0 192.168.1.1 46898 } <nil>
    2024/10/19 19:19:02 {0 192.168.1.1 8080 } <nil>
    2024/10/19 19:19:02 Remote Addr 192.168.1.1:46898, Local Address 192.168.1.1:8080
    2024/10/19 19:19:02 Received message on Server: Hello, TCP Server!
    2024/10/19 19:19:02 Received response: Hey TCP Client
    ```

2. Userspace Http client and server (https://viveksb007.github.io/2024/10/gvisor-userspace-http-server-client/)

    `go run cmd/userspace-http/main.go`
    ```
    2024/10/27 20:53:11 Starting http server on port 8080
    2024/10/27 20:53:11 {0 192.168.1.1 35597 } <nil>
    2024/10/27 20:53:11 {0 192.168.1.1 8080 } <nil>
    2024/10/27 20:53:11 TCP connection to 192.168.1.1:8080 is successful
    2024/10/27 20:53:11 Received request from client 192.168.1.1:35597
    2024/10/27 20:53:11 Received response: yo man
    2024/10/27 20:53:11 {0 192.168.1.1 35598 } <nil>
    2024/10/27 20:53:11 {0 192.168.1.1 8080 } <nil>
    2024/10/27 20:53:11 TCP connection to 192.168.1.1:8080 is successful
    2024/10/27 20:53:11 Received request from client 192.168.1.1:35598
    2024/10/27 20:53:11 Received response: yo man
    2024/10/27 20:53:11 {0 192.168.1.1 35599 } <nil>
    2024/10/27 20:53:11 {0 192.168.1.1 8080 } <nil>
    2024/10/27 20:53:11 TCP connection to 192.168.1.1:8080 is successful
    2024/10/27 20:53:11 Received request from client 192.168.1.1:35599
    2024/10/27 20:53:11 Received response: yo man
    ```

3. Userspace Http client communicating to Internet (https://viveksb007.github.io/2024/11/gvisor-userspace-routing-to-internet/)

    `go run cmd/userspace-tcpclient/main.go -domain google.com`
    ```
    I1121 21:47:18.447863   48963 sniffer.go:378] send tcp 192.168.1.2:61026 -> 172.253.62.100:80 len:0 id:7622 flags:  S       seqnum: 3215243184 ack: 0 win: 29184 xsum:0xbc2e options: {MSS:1460 WS:7 TS:true TSVal:3731565394 TSEcr:0 SACKPermitted:false Flags:        }
    I1121 21:47:18.467162   48963 sniffer.go:378] recv tcp 172.253.62.100:80 -> 192.168.1.2:61026 len:0 id:f42f flags:  S  A    seqnum: 3303876076 ack: 3215243185 win: 29184 xsum:0x72e2 options: {MSS:1460 WS:7 TS:true TSVal:592919306 TSEcr:3731565394 SACKPermitted:false Flags:        }
    I1121 21:47:18.467348   48963 sniffer.go:378] send tcp 192.168.1.2:61026 -> 172.253.62.100:80 len:0 id:7623 flags:     A    seqnum: 3215243185 ack: 3303876077 win: 228 xsum:0xfb7 options: {TS:true TSVal:3731565413 TSEcr:592919306 SACKBlocks:[]}
    2024/11/21 21:47:18 {1 192.168.1.2 61026 } <nil>
    2024/11/21 21:47:18 {1 172.253.62.100 80 } <nil>
    2024/11/21 21:47:18 TCP connection to 172.253.62.100:80 is successful
    I1121 21:47:18.467487   48963 sniffer.go:378] send tcp 192.168.1.2:61026 -> 172.253.62.100:80 len:68 id:7624 flags:    PA    seqnum: 3215243185 ack: 3303876077 win: 4096 xsum:0x4646 options: {TS:true TSVal:3731565413 TSEcr:592919306 SACKBlocks:[]}
    I1121 21:47:18.467747   48963 sniffer.go:378] recv tcp 172.253.62.100:80 -> 192.168.1.2:61026 len:0 id:0000 flags:     A    seqnum: 3303876077 ack: 3215243253 win: 227 xsum:0xf74 options: {TS:true TSVal:592919306 TSEcr:3731565413 SACKBlocks:[]}
    I1121 21:47:18.493076   48963 sniffer.go:378] recv tcp 172.253.62.100:80 -> 192.168.1.2:61026 len:773 id:0000 flags:    PA    seqnum: 3303876077 ack: 3215243253 win: 4096 xsum:0x53 options: {TS:true TSVal:592919332 TSEcr:3731565413 SACKBlocks:[]}
    ```