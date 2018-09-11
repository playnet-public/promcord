###################### //S/M Makefile ######################
#
# Edit this file with care, as it is also being used by our CI/CD Pipeline
# For usage information check README.md
#
# Parts of this makefile are based upon github.com/kolide/kit
#

export NAME		:= promcord
export REPO		:= playnet-public
export GIT_HOST	:= github.com
export REGISTRY	?= eu.gcr.io
export PROJECT		?= playnet-gce

export PATH 		:= $(GOPATH)/bin:$(PATH)
export BAZEL_PYTHON=python2.7

export NAMESPACE	?= $(NAME)-dev

-include .env

include helpers/make_version

.PHONY: build

### MAIN STEPS ###

all: test install run

# install required tools and dependencies
deps:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/golang/lint/golint
	go get -u github.com/haya14busa/goverage
	go get -u github.com/kisielk/errcheck
	go get -u github.com/maxbrunsfeld/counterfeiter
	go get -u github.com/onsi/ginkgo/ginkgo
	go get -u github.com/onsi/gomega
	go get -u github.com/schrej/godacov
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/bborbe/teamvault-utils/cmd/teamvault-config-parser

updateDebugger:
	wget -O files/go-cloud-debug https://storage.googleapis.com/cloud-debugger/compute-go/go-cloud-debug
	chmod 0755 files/go-cloud-debug

# test entire repo
gotest:
	@go test -cover -race $(shell go list ./... | grep -v /vendor/)

test:
	@go get github.com/onsi/ginkgo/ginkgo
	@ginkgo -r -race

gazelle:
	bazel run //:gazelle

build: gazelle
	bazel build --workspace_status_command=./helpers/status.sh //...

docker: gazelle
	bazel run --force_python=py2 --workspace_status_command=./helpers/status.sh $(CMD):image -- --norun

push: gazelle
	bazel run --workspace_status_command=./helpers/status.sh //:push

run: gazelle
	bazel run \
	--workspace_status_command=./helpers/status.sh $(CMD):bin

# update the secret located inside k8s/secret.yaml for NAMESPACE
secret: parser
	cat k8s/secret.yaml | teamvault-config-parser \
	-teamvault-config="$(TEAMVAULT)" \
	-logtostderr \
	-v=2 | kubectl apply --namespace=$(NAMESPACE) -f -

kube: gazelle secret
	bazel run \
	--workspace_status_command=./helpers/status.sh $(CMD)

# install passed in tool project
install:
	GOBIN=$(GOPATH)/bin go install cmd/$(NAME)/*.go

# run specified tool from code
dev: generate
	@go run -ldflags $(KIT_VERSION) cmd/$(NAME)/*.go \
	-debug

# format entire repo (excluding vendor)
format:
	@go get golang.org/x/tools/cmd/goimports
	@find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w "{}" +
	@find . -type f -name '*.go' -not -path './vendor/*' -exec goimports -w "{}" +

clean:
	bazel clean

version:
	@echo $(VERSION)


# go quality checks
check: format lint vet

# vet entire repo (excluding vendor)
vet:
	@go vet $(shell go list ./... | grep -v /vendor/)

# lint entire repo (excluding vendor)
lint:
	@go get github.com/golang/lint/golint
	@golint -min_confidence 1 $(shell go list ./... | grep -v /vendor/)

# errcheck entire repo (excluding vendor)
errcheck:
	@go get github.com/kisielk/errcheck
	@errcheck -ignore '(Close|Write)' $(shell go list ./... | grep -v /vendor/)

cover:
	@go get github.com/haya14busa/goverage
	@go get github.com/schrej/godacov
	goverage -v -coverprofile=coverage.out $(shell go list ./... | grep -v /vendor/)

generate:
	@go get github.com/maxbrunsfeld/counterfeiter
	@go generate ./...

parser:
	@go get github.com/bborbe/teamvault-utils/cmd/teamvault-config-parser