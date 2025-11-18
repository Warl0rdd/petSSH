package utils

import (
	"golang.org/x/crypto/ssh"
	"os"
)

// FileExists reports whether path exists (dir or file)
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsRegularFile checks whether path exists and is a regular file
func IsRegularFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

// LoadAuthorizedKeys reads authorized keys file and returns map[marshal]bool
func LoadAuthorizedKeys(path string) (map[string]bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	m := make(map[string]bool)
	rest := b
	for len(rest) > 0 {
		pubKey, _, _, r, err := ssh.ParseAuthorizedKey(rest)
		if err != nil {
			return nil, err
		}
		m[string(pubKey.Marshal())] = true
		rest = r
	}
	return m, nil
}

// lLoadPrivateKey parses the private key file and returns ssh.Signer
func LoadPrivateKey(path string) (ssh.Signer, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(b)
}
