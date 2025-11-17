package main

import (
	"crypto"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s <output directory>\n", os.Args[0])
	}

	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}

	p, err := ssh.MarshalPrivateKey(crypto.PrivateKey(priv), "")
	if err != nil {
		panic(err)
	}

	privateKeyPem := pem.EncodeToMemory(p)
	privateKeyString := string(privateKeyPem)
	publicKey, err := ssh.NewPublicKey(pub)
	if err != nil {
		panic(err)
	}

	publicKeyString := "ssh-ed25519" + " " + base64.StdEncoding.EncodeToString(publicKey.Marshal())

	err = os.MkdirAll(os.Args[1], 0700)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(os.Args[1]+"/ssh_host_ed25519", []byte(privateKeyString), 0600)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(os.Args[1]+"/ssh_host_ed25519.pub", []byte(publicKeyString), 0700)
	if err != nil {
		panic(err)
	}
}
