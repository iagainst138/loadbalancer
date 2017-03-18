build:
	go build -o loadbalancer config.go hash.go leastconn.go log.go main.go proxy.go roundrobin.go
