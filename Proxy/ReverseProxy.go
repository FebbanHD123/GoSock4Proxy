package Proxy

import (
	"fmt"
	"net"
)

type UpstreamBridge struct {
	address string
	port    int
	conn    *net.TCPConn
}

func Start(remoteAddress string, remotePort int) {
	fmt.Println("Starting ")
	address, err := net.ResolveTCPAddr("tcp", "localhost:25565")
	if err != nil {
		panic(err)
	}
	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	fmt.Println("Listening on", address)
	for {
		connection, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}
		connection.SetNoDelay(true)
		connection.SetKeepAlive(true)

		go handleConnection(connection, UpstreamBridge{
			address: remoteAddress,
			port:    remotePort,
		})
	}
}

func handleConnection(conn *net.TCPConn, upstream UpstreamBridge) {
	fmt.Println("Handle new connection:", conn.RemoteAddr())
	err := upstream.openConnection()
	if err != nil {
		conn.Close()
		fmt.Println(err)
		return
	}
	fmt.Println("Connected to upstream")

	go func() {
		for {
			buf := make([]byte, 1024)
			size, err := upstream.conn.Read(buf)
			if err != nil {
				fmt.Println("Closed connection 1:", err.Error())
				break
			}
			buf = buf[:size]
			if size > 0 {
				conn.Write(buf)
			}
		}
	}()

	for {
		buf := make([]byte, 1024)
		size, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Closed connection 2:", err.Error())
			break
		}
		buf = buf[:size]
		if size > 0 {
			upstream.conn.Write(buf)
		}
	}
}

func (upstream *UpstreamBridge) openConnection() error {
	proxy := CreateSockProxy("202.21.112.172", 1080, upstream.address, upstream.port)
	conn, err := proxy.Connect()
	if err != nil {
		return err
	}
	conn.SetNoDelay(true)
	conn.SetKeepAlive(true)
	upstream.conn = conn
	return nil
}

func main() {
	Start("gommehd.net", 25565)
}
