package Proxy

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type SocksProxy struct {
	proxyIp, destinationIp     string
	proxyPort, destinationPort int
}

func CreateSockProxy(proxyIp string, proxyPort int, destinationIp string, destinationPort int) SocksProxy {
	return SocksProxy{
		proxyPort:       proxyPort,
		proxyIp:         proxyIp,
		destinationPort: destinationPort,
		destinationIp:   destinationIp,
	}
}

func (p *SocksProxy) Connect() (*net.TCPConn, error) {
	fmt.Println("Connecting to proxy")
	address, err := net.ResolveTCPAddr("tcp", p.proxyIp+":"+strconv.Itoa(p.proxyPort))
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, address)
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to proxy server")
	fmt.Println("Send handshake")

	//protocol: https://de.wikipedia.org/wiki/SOCKS
	//send "handshake" packet
	var packet = []byte{
		//sock4 versoin
		0x04,
		//new tcp connection
		0x01,
	}

	portBytes := p.getDestinationPortAsByteArray()
	for i := 0; i < len(portBytes); i++ {
		packet = append(packet, portBytes[i])
	}

	//destination ip
	ipBytes := p.getDestinationIpAsByteArray()
	for i := 0; i < len(ipBytes); i++ {
		packet = append(packet, ipBytes[i])
	}

	//end
	packet = append(packet, 0x00)
	conn.Write(packet)

	for {
		buf := make([]byte, 1024)
		size, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Closed connection:", err.Error())
			break
		}
		buf = buf[:size]
		if len(buf) > 1 {
			//answer
			code := buf[1]
			if code == 0x5A {
				return conn, nil
			} else if code == 0x5B {
				fmt.Println("Proxy accepted connection")
				return conn, errors.New("Request rejected or failed")
			} else if code == 0x5C {
				return conn, errors.New("Request failed because the client does not execute identd (or is not reachable from the server)")
			} else if code == 0x5D {
				return conn, errors.New("Request failed because identd could not confirm the ID")
			}
			break
		}
	}
	return conn, errors.New("Unecpeted")
}

func (p SocksProxy) getDestinationIpAsByteArray() []byte {
	strs := strings.Split(p.destinationIp, ".")
	bytes := make([]byte, 0)
	for i := 0; i < len(strs); i++ {
		intValue, _ := strconv.Atoi(strs[i])
		bytes = append(bytes, byte(intValue))
	}
	return bytes
}

func (p SocksProxy) getDestinationPortAsByteArray() []byte {
	b := make([]byte, 2)
	b[0], b[1] = byte(p.destinationPort>>8), byte(p.destinationPort&255)
	return b
}
