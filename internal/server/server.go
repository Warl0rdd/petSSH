package server

import (
	"errors"
	"log/slog"
	"net"
	"strings"
)

type Server struct {
	Addr string
	ln   net.Listener
}

func New(addr string) (*Server, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, errors.New("empty address")
	}

	return &Server{
		Addr: addr,
	}, nil
}

func (s *Server) ListenAndServe(done <-chan struct{}) error {
	slog.Info("starting server", "addr", s.Addr)
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	s.ln = ln
	defer ln.Close()
	slog.Info("server started", "addr", s.Addr)

	errCh := make(chan error, 1)

	go func() {
		for {
			conn, err := s.ln.Accept()
			slog.Info("connection accepted", "addr", conn.RemoteAddr().String())

			if err != nil {
				select {
				case <-done:
					errCh <- nil
				default:
					errCh <- err
				}
				return
			}

			go func(c net.Conn) {
				// s.handleConn(c)
			}(conn)
		}
	}()

	select {
	case <-done:
		_ = s.ln.Close()
		return nil
	case err := <-errCh:
		_ = s.ln.Close()
		if err != nil {
			return err
		}
		return nil
	}

}
