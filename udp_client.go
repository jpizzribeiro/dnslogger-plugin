package dnslogger

import (
	"fmt"
	"net"
)

// UDPClient gerencia a conexão UDP
type UDPClient struct {
	Conn       *net.UDPConn
	SocketAddr string
}

// NewUDPClient cria um novo cliente UDP
func NewUDPClient(socketAddr string) (*UDPClient, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", socketAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to UDP socket: %w", err)
	}

	_, err = conn.Write([]byte("connected on udp socket\n"))
	if err != nil {
		return nil, err
	}

	return &UDPClient{
		Conn:       conn,
		SocketAddr: socketAddr,
	}, nil
}

// Send envia uma mensagem via UDP
func (c *UDPClient) Send(message string) error {
	_, err := c.Conn.Write([]byte(message))
	return err
}

// Close fecha a conexão UDP
func (c *UDPClient) Close() error {
	return c.Conn.Close()
}
