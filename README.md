# loadbalancer

A basic TCP loadbalancer that can use round robin, hash, and least number of connections mechanisms for proxying to backends. Can also be used for TLS termination. See the sample_configs directory for sample configs.


## Building

go build -o loadbalancer config.go hash.go leastconn.go log.go main.go proxy.go roundrobin.go


## Running

#### Create TLS certs
./make-cert.sh

#### Start some backends
./create-backends.sh

#### Run the loadbalancer
./loadbalancer -config sample_configs/config.json

#### Make a connection
curl -v http://localhost:9090

#### Make a connection using TLS
curl -v --cacert certs/server.pem https://localhost:8080
