package snmplib

import (
	"math/rand"
	"net"
	"time"
)

// TrapHandler interface.
type TrapHandler interface {
	OnError(addr net.Addr, err error)
	OnTrap(addr net.Addr, trap Trap)
}

// TrapServer object.
type TrapServer struct {
	PacketSize int
	IPAddress  net.UDPAddr
	Port       int
	Conn       *net.UDPConn
	Users      []V3user
}

// NewTrapServer creates a new TrapServer object.
func NewTrapServer(ip string, port int) (TrapServer, error) {
	rand.Seed(0)
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(ip),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return TrapServer{}, err
	}
	return TrapServer{PacketSize: 3000, IPAddress: addr, Port: port, Conn: conn}, nil
}

// ListenAndServe starts the listen loop and will pause execution until server is shut down.
func (s *TrapServer) ListenAndServe(handler TrapHandler) {
	server := NewSNMPOnConn("", "", SNMPv3, 2*time.Second, 5, s.Conn)
	defer server.Close()

	server.TrapUsers = s.Users

	packet := make([]byte, s.PacketSize)
	for {
		_, addr, err := s.Conn.ReadFromUDP(packet)
		if err != nil {
			handler.OnError(addr, err)
			continue
		}

		trap, err := server.ParseTrap(packet)
		if err != nil {
			handler.OnError(addr, err)
			continue
		}
		if trap.Address == "" {
			trap.Address = addr.String()
		}

		handler.OnTrap(addr, trap)
	}
}
