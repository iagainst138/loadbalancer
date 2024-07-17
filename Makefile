SHELL=/bin/bash
#COVERAGE_DIR=$(abspath $(lastword $(MAKEFILE_LIST)))
ROOT_DIR=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
COVERAGE_DIR=${ROOT_DIR}/coverage


certs:
	./make-cert.sh


build_backend: certs
	(cd lb && go build ./cmd/backend/)


build_lb: certs
	(cd lb && go build -o lb-main ./cmd/lb/)


run_backends: build_backend
	./lb/backend


run_lb: build_lb
	./lb/lb-main -config sample_configs/config.yml -start-http -server-addr 127.0.0.1:9444


clean:
	rm -rf certs lb/backend lb/lb-main coverage


unit_test:
	@mkdir -p ${COVERAGE_DIR}
	@(cd lb && go test lb -v -cover -coverprofile=${COVERAGE_DIR}/coverage.out .)
	@(cd lb && go tool cover -html ${COVERAGE_DIR}/coverage.out -o ${COVERAGE_DIR}/coverage.html)


test_client:
	@curl -s 127.0.0.1:8081


loop_test_client:
	watch -t -d curl -s 127.0.0.1:8081
