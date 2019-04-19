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
	CGO_ENABLED=0 go build -o bin/static-pod-controller
clean:
	rm -rf bin/ dist/

image:
	docker build -f package/Dockerfile -t cnrancher/static-pod-controller .
	docker push cnrancher/static-pod-controller
