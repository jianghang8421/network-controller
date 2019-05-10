TARGETS := $(shell ls scripts)

.dapper:
	@echo Downloading dapper
	@curl -sL https://releases.rancher.com/dapper/latest/dapper-`uname -s`-`uname -m` > .dapper.tmp
	@@chmod +x .dapper.tmp
	@./.dapper.tmp -v
	@mv .dapper.tmp .dapper

$(TARGETS): .dapper
	./.dapper $@

trash: .dapper
	./.dapper -m bind trash

trash-keep: .dapper
	./.dapper -m bind trash -k

deps: trash

.DEFAULT_GOAL := ci

.PHONY: $(TARGETS) dev clean

dev:
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/network-controller
clean:
	rm -rf bin/ dist/

image:
	docker build -f package/Dockerfile -t wardenlym/network-controller:v0.2.1 .
	docker push wardenlym/network-controller:v0.2.1

poc:
	cat ./artifacts/multus-daemonset.yml \
		./artifacts/network-cni-daemonset.yml \
		./artifacts/flannel-daemonset.yml \
		./artifacts/network-controller.yml > ./artifacts/poc/macvlan-network.yml