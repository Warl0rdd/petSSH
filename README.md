# petSSH
SSH daemon implemented in golang. Developed as a pet project, do not use in production.

## Host key generation

Host keys are generated with [ed25519](https://en.wikipedia.org/wiki/Ed25519). You can either generate them with ssh keygen: 

```bash
ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519
```

Or use the provided utility:

```bash
go run cmd/genHostKey/main.go ~/.ssh
```

## Usage

```bash
go build ./cmd/sshd
sudo ./sshd -a=:2323 -keyDir=~/.ssh -ak=~/.ssh/authorized_keys
```

Important: sshd is required to start as root as it needs access to PAM API. Processes' privileges of each user's connection are dropped according to one's privileges.