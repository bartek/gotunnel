package tunnel

import (
	"os"

	"golang.org/x/crypto/ssh"
)

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
