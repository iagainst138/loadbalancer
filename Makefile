build:
	go build -i -o loadbalancer config.go hash.go leastconn.go log.go main.go proxy.go roundrobin.go mgmt.go resources.go
