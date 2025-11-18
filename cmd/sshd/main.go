package main

import (
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"petssh/internal/server"
	"petssh/internal/utils"
	"runtime"
	"syscall"
)

func main() {
	// Platform check
	if runtime.GOOS != "linux" {
		log.Fatalf("unsupported OS: %s, only linux is supported", runtime.GOOS)
	}
	runtime.GOMAXPROCS(runtime.NumCPU())

	home := os.Getenv("HOME")
	if home == "" {
		log.Fatalf("HOME env is not set")
	}

	defaultHostKeyDir := filepath.Join(home, ".ssh")

	var (
		addr                = flag.String("a", ":22", "server address")
		hostKeyDir          = flag.String("keyDir", defaultHostKeyDir, "host key directory")
		authorizedKeysFile  = flag.String("ak", filepath.Join(defaultHostKeyDir, "authorized_keys"), "authorized_keys file")
		passwordAuthEnabled = flag.Bool("p", false, "enable password auth")
	)
	flag.Parse()

	// Validate required files
	privPath := filepath.Join(*hostKeyDir, "ssh_host_ed25519")

	if !utils.FileExists(privPath) {
		log.Fatalf("missing private host key: %s", privPath)
	}
	if !utils.IsRegularFile(*authorizedKeysFile) {
		log.Fatalf("authorized keys file does not exist or is not a regular file: %s", *authorizedKeysFile)
	}

	// Load authorized keys
	authorizedKeysMap, err := utils.LoadAuthorizedKeys(*authorizedKeysFile)
	if err != nil {
		log.Fatalf("failed to load authorized keys: %v", err)
	}

	// Load host private key
	private, err := utils.LoadPrivateKey(privPath)
	if err != nil {
		log.Fatalf("failed to load private key: %v", err)
	}

	// Prepare ssh.ServerConfig
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

	// Create server
	srv, err := server.New(*addr, config)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigCh
		log.Println("shutdown signal received")
		close(done)
	}()

	log.Printf("starting server on %s (password auth: %t)", *addr, *passwordAuthEnabled)
	if err := srv.ListenAndServe(done); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
	log.Println("server stopped")
}
