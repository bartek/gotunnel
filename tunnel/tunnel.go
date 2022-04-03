package tunnel

// REVIEW: https://ixday.github.io/post/golang_ssh_tunneling/

import (
	"fmt"
	"io"
	"net"

	"golang.org/x/crypto/ssh"
)

// Endpoint contains information about the target address
type Endpoint struct {
	Address string
	User    string
}

func (e Endpoint) String() string {
	return e.Address
}

func NewEndpoint(s string) *Endpoint {
	// FIXME: Find @ to check for user and split.
	return &Endpoint{
		Address: s,
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

// awaitConnection awaits a connection to listener via Accept
// When the connection is made the resulting net.Conn is sent on the channel
func awaitConnection(listener net.Listener, c chan net.Conn) {
	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Accepting error: ", err)
		return
	}
	c <- conn
}

func (s *SSHTunnel) Start() error {
	listener, err := net.Listen("tcp", s.Local.String())
	if err != nil {
		return err
	}

	c := make(chan net.Conn)

	for !s.closed {
		go awaitConnection(listener, c)

		select {
		case <-s.close:
			s.closed = true
		case conn := <-c:
			go func(conn net.Conn) {
				// Now forward the connection by sshing into the target and assigning the
				// remote.
				serverConn, err := ssh.Dial("tcp", s.Target.String(), s.Config)
				if err != nil {
					return
				}

				fmt.Println("connected to", s.Target.String())

				// Dial the remote endpoint using server connection
				remoteConn, err := serverConn.Dial("tcp", s.Remote.String())
				if err != nil {
					return
				}

				copyConn := func(writer, reader net.Conn) {
					_, err := io.Copy(writer, reader)
					if err != nil {
						fmt.Printf("tunnel copy error: %s", err)
					}
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
	return &SSHTunnel{
		Target: NewEndpoint(target),
		Local:  NewEndpoint(local),
		Remote: NewEndpoint(destination),

		Config: &ssh.ClientConfig{
			User: "ubuntu",
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
