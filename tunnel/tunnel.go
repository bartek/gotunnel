package tunnel

import (
	"fmt"
	"io"
	"net"
	"strings"

	"golang.org/x/crypto/ssh"
)

// Endpoint contains information about the target address
type Endpoint struct {
	Address string
	User    string
}

// String implements Stringer
func (e Endpoint) String() string {
	builder := strings.Builder{}
	builder.WriteString(e.User)
	if e.User != "" {
		builder.WriteString("@")
	}
	builder.WriteString(e.Address)
	return builder.String()
}

func NewEndpoint(s string) *Endpoint {
	// Check for @ to identify username.
	idx := strings.Index(s, "@")
	var user string
	if idx > -1 {
		user = s[:idx]
		s = s[idx+1:]
	}

	return &Endpoint{
		Address: s,
		User:    user,
	}
}

type SSHTunnel struct {
	// Local is the local endpoint
	Local *Endpoint

	// Remote is the remote endpoint
	Remote *Endpoint

	// Target is the target server to tunnel through
	Target *Endpoint

	// Config contains SSH client configuration, in particular the
	// authentication method, username.
	Config *ssh.ClientConfig

	closed bool
	close  chan struct{}
}

func copyConn(writer, reader net.Conn) {
	_, err := io.Copy(writer, reader)
	if err != nil {
		fmt.Printf("tunnel copy error: %s", err)
	}
}

func (s *SSHTunnel) Start() error {
	listener, err := net.Listen("tcp", s.Local.Address)
	if err != nil {
		return err
	}

	for !s.closed {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		select {
		case <-s.close:
			s.closed = true
		default:
			go func(conn net.Conn) {
				// Now forward the connection by sshing into the target and assigning the
				// remote.
				serverConn, err := ssh.Dial("tcp", s.Target.Address, s.Config)
				if err != nil {
					return
				}

				fmt.Println("connected to", s.Target.Address)

				// Dial the remote endpoint using server connection
				remoteConn, err := serverConn.Dial("tcp", s.Remote.Address)
				if err != nil {
					return
				}

				go copyConn(conn, remoteConn)
				go copyConn(remoteConn, conn)
			}(conn)
		}

	}

	return nil
}

// Close closes the SSH tunnel
func (s *SSHTunnel) Close() {
	s.close <- struct{}{}
}

func New(target string, auth ssh.AuthMethod, local string, destination string) *SSHTunnel {
	t := NewEndpoint(target)

	return &SSHTunnel{
		Target: t,
		Local:  NewEndpoint(local),
		Remote: NewEndpoint(destination),

		Config: &ssh.ClientConfig{
			User: t.User,
			Auth: []ssh.AuthMethod{auth},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				// Always accept key. We probably shouldn't do this, but for the
				// sake of simplicity...
				return nil
			},
		},

		close: make(chan struct{}),
	}
}
