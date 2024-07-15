SHELL=/bin/bash


certs:
	./make-cert.sh


build_backend: certs
	(cd lb && go build ./cmd/backend/)


build_lb: certs
	(cd lb && go build -o lb-main ./cmd/lb/)


run_backends: build_backend
	./lb/backend


run_lb: build_lb
	./lb/lb-main -config sample_configs/config.json


clean:
	rm -rf certs lb/backend


test_client:
	@curl -s 127.0.0.1:8081


loop_test_client:
	watch -t -d curl -s 127.0.0.1:8081
