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
./sshd -a=:2323 -keyDir=~/.ssh -ak=~/.ssh/authorized_keys
```