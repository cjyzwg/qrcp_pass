package server

import (
	"net"
	"time"
)

//TcpKeepAliveListener is a struct
type TcpKeepAliveListener struct {
	*net.TCPListener
}

//Accept is a function
func (ln TcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, err
}
