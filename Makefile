APP_DIR_LIST=$(shell go list ./... | grep -v /vendor/)
GOBIN=$(GOPATH)/bin

build:
	CGO_ENABLED=0 go install -tags netgo ${APP_DIR_LIST}
	go fmt $(APP_DIR_LIST)

bin/govendor: verify_gopath
	go get -v -u github.com/kardianos/govendor

deps_fetch_specific: bin/govendor
	@if [ "$(DEP_URL)" = "" ]; then\
		echo "DEP_URL not set. Run this comand as follow:";\
		echo " make deps_fetch_specific DEP_URL=github.com/nu7hatch/gouuid";\
	exit 1 ;\
	fi
	@echo "Fetching specific dependency in newest versions"
	$(GOBIN)/govendor fetch -v $(DEP_URL)

deps_update_tap: verify_gopath
	$(GOBIN)/govendor update github.com/trustedanalytics/...
	rm -Rf vendor/github.com/trustedanalytics/tap-ceph-broker
	@echo "Done"

verify_gopath:
	@if [ -z "$(GOPATH)" ] || [ "$(GOPATH)" = "" ]; then\
		echo "GOPATH not set. You need to set GOPATH before run this command";\
		exit 1 ;\
	fi

gometalinter: verify_gopath
	$(GOBIN)/gometalinter `find -type d | grep -v vendor | grep -v .git` --deadline=10s

tests: verify_gopath
	go test --cover $(APP_DIR_LIST)

prepare_dirs:
	mkdir -p ./temp/src/github.com/trustedanalytics/tap-ceph-broker
	$(eval REPOFILES=$(shell pwd)/*)
	ln -sf $(REPOFILES) temp/src/github.com/trustedanalytics/tap-ceph-broker

build_anywhere: prepare_dirs
	$(eval GOPATH=$(shell cd ./temp; pwd))
	$(eval APP_DIR_LIST=$(shell GOPATH=$(GOPATH) go list ./temp/src/github.com/trustedanalytics/tap-ceph-broker/... | grep -v /vendor/))
	GOPATH=$(GOPATH) CGO_ENABLED=0 go build -tags netgo $(APP_DIR_LIST)
	rm -Rf application && mkdir application
	cp -RL ./tap-ceph-broker ./application/tap-ceph-broker
	rm -Rf ./temp

test: verify_gopath
	go test --cover $(APP_DIR_LIST)
