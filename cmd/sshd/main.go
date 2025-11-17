package main

import (
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"os/signal"
	"petssh/internal/server"
	"runtime"
	"syscall"
)

func main() {
	// Verify that the OS is linux and process command line flags
	if runtime.GOOS != "linux" {
		log.Fatalf("unsupported OS: %s, only linux is supported", runtime.GOOS)
	}

	defaultHostKeyDir := os.Getenv("HOME") + "/.ssh"

	var (
		addr                = flag.String("a", ":80", "server address")
		hostKeyDir          = flag.String("keyDir", defaultHostKeyDir, "host key directory")
		authorizedKeysDir   = flag.String("ak", defaultHostKeyDir+"/authorized_keys", "authorized keys file")
		passwordAuthEnabled = flag.Bool("p", false, "enable password auth")
	)

	flag.Parse()

	// Verify required files existence
	if _, err := os.Stat(*hostKeyDir + "/ssh_host_ed25519"); err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(*hostKeyDir + "/ssh_host_ed25519.pub"); err != nil {
		log.Fatal(err)
	}

	if info, err := os.Stat(*authorizedKeysDir); err != nil || !info.Mode().IsRegular() {
		log.Fatal("authorized keys file does not exist or is not a regular file")
	}

	authorizedKeysBytes, err := os.ReadFile("authorized_keys")
	if err != nil {
		log.Fatalf("Failed to load authorized_keys, err: %v", err)
	}

	// Parse authorized keys
	authorizedKeysMap := map[string]bool{}
	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			log.Fatal(err)
		}

		authorizedKeysMap[string(pubKey.Marshal())] = true
		authorizedKeysBytes = rest
	}

	// Parse host private key

	privateBytes, err := os.ReadFile(*hostKeyDir + "/ssh_host_ed25519")
	if err != nil {
		log.Fatal("Failed to load private key: ", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}

	// Setup ssh config
	config := &ssh.ServerConfig{
		MaxAuthTries: 3,
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			if authorizedKeysMap[string(pubKey.Marshal())] {
				return &ssh.Permissions{
					Extensions: map[string]string{
						"pubkey-fp": ssh.FingerprintSHA256(pubKey),
					},
				}, nil
			}
			return nil, fmt.Errorf("unknown public key for %q", c.User())
		},
	}

	config.AddHostKey(private)

	if *passwordAuthEnabled {
		// TODO implement
	}

	// Setup server
	s, err := server.New(*addr, config)

	if err != nil {
		log.Fatal(err)
	}

	if err := s.ListenAndServe(nil); err != nil {
		log.Fatal(err)
	}

	done := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigCh
		fmt.Println("Shutting down...")
		close(done)
	}()

	if err := s.ListenAndServe(done); err != nil {
		log.Fatal(err)
	}
}
