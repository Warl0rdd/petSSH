package server

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"log/slog"
	"net"
	"strings"
)

type Server struct {
	Addr      string
	sshConfig *ssh.ServerConfig
	ln        net.Listener
}

func New(addr string, config *ssh.ServerConfig) (*Server, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, errors.New("empty address")
	}

	return &Server{
		Addr:      addr,
		sshConfig: config,
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
				// TODO: closure for graceful shutdown handling
				sshConn, chans, reqs, err := ssh.NewServerConn(c, s.sshConfig)
				if err != nil {
					slog.Info("failed handshake", "err", err, "addr", conn.RemoteAddr().String())
					return
				}
				s.handleConn(sshConn, chans, reqs)
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

func (s *Server) handleConn(conn *ssh.ServerConn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request) {
	go ssh.DiscardRequests(reqs)
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Fatalf("Could not accept channel: %v", err)
		}

		go func(in <-chan *ssh.Request) {
			for req := range in {
				_ = req.Reply(req.Type == "shell", nil)
			}
		}(requests)

		term := terminal.NewTerminal(channel, "> ")

		go func() {
			defer func() {
				_ = channel.Close()
				_ = conn.Close()
			}()
			for {
				line, err := term.ReadLine()
				if err != nil {
					break
				}
				fmt.Println(line)
			}
		}()
	}
}
