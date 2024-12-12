package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"sync"
)

type ConnectionPool struct {
	mu       sync.Mutex
	conns    []*smtp.Client
	host     string
	port     string
	maxConns int
}

func NewConnectionPool(host, port string, maxConns int) *ConnectionPool {
	return &ConnectionPool{
		host:     host,
		port:     port,
		maxConns: maxConns,
	}
}

func (p *ConnectionPool) Get() (*smtp.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.conns) > 0 {
		conn := p.conns[len(p.conns)-1]
		p.conns = p.conns[:len(p.conns)-1]
		return conn, nil
	}

	addr := fmt.Sprintf("%s:%s", p.host, p.port)
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		return nil, err
	}
	return smtp.NewClient(conn, p.host)
}

func (p *ConnectionPool) Put(conn *smtp.Client) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.conns) < p.maxConns {
		p.conns = append(p.conns, conn)
	} else {
		conn.Close()
	}
}
