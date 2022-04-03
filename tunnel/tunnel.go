package tunnel

// REVIEW: https://ixday.github.io/post/golang_ssh_tunneling/

import (
	"fmt"
	"io"
	"net"
	"os"

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

func awaitConnection(listener net.Listener, c chan net.Conn) {
	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Accepting error: ", err)
		return
	}
	fmt.Println("returned conn")
	c <- conn
}

func (s *SSHTunnel) Start() error {
	fmt.Println(s.Local.String())
	listener, err := net.Listen("tcp", s.Local.String())
	if err != nil {
		return err
	}

	for !s.closed {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accepting error: ", err)
		}

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

		<-s.close
		fmt.Println("closed")
		s.closed = true

	}

	return nil
}

// Close closes the SSH tunnel
func (s *SSHTunnel) Close() {
	s.close <- struct{}{}
}

func PEMFile(path string) ssh.AuthMethod {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(b)
	if err != nil {
		return nil
	}

	return ssh.PublicKeys(key)
}

func New(target string, auth ssh.AuthMethod, local string, destination string) *SSHTunnel {

	return &SSHTunnel{
		Target: NewEndpoint(target),
		Local:  NewEndpoint(local),
		Remote: NewEndpoint(destination),

		Config: &ssh.ClientConfig{
			Auth: []ssh.AuthMethod{auth},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				// Always accept key.
				return nil
			},
		},

		close: make(chan struct{}),
	}
}
