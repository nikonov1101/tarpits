default: ssh http

.PHONY:ssh
ssh:
	GOOS=linux GOARCH=amd64 go build -o ssh_tarpit ./ssh/main.go

.PHONY: http
http:
	GOOS=linux GOARCH=amd64 go build -o http_tarpit ./http/main.go
