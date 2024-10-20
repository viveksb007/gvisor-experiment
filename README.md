## gvisor experiments

1. Userspace TCP client and server (https://viveksb007.github.io/2024/10/gvisor-userspace-tcp-server-client/)

    `go run cmd/userspace-tcpip/main.go`
    ```
    viveksb007@Viveks-MacBook-Air-2  ~/workplace/gvisor-experiment  go run cmd/userspace-tcpip/main.go
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

