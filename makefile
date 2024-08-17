default: ssh

.PHONY:ssh
ssh:
	GOOS=linux GOARCH=amd64 go build -o ssh_tarpit ./ssh/main.go
