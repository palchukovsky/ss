# Copyright 2021, the SS project owners. All rights reserved.
# Please see the OWNERS and LICENSE files for details.

.PHONY: all\
	install-env \
	lint \
	mock \
.DEFAULT_GOAL = all

GO_VER = 1.17
# golangci-lint will be installed inside container during building, local
# copy has to be installed by:
#   go install github.com/golangci/golangci-lint/cmd/golangci-lint@v${GOLANGCI_VER}
# or
#   brew update && brew upgrade golangci-lint
GOLANGCI_VER = 1.42.0

ORGANIZATION = palchukovsky
CODE_REPO = github.com/${ORGANIZATION}/ss


define echo_start
	@echo ================================================================================
	@echo :
	@echo : START: $(@)
	@echo :
endef
define echo_success
	@echo :
	@echo : SUCCESS: $(@)
	@echo :
	@echo ================================================================================
endef

define make_target
	$(MAKE) -f ./Makefile ${1}
endef

help: ## Show this help.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' ${MAKEFILE_LIST} | sort | awk 'BEGIN {FS = ":.*?## "};	{printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'


all:
	@$(call echo_start)
	go mod download
	$(call make_target,install-env)
	$(call make_target,lint)
	$(call make_target,mock)
	go test -timeout 15s -v -coverprofile=coverage.txt -covermode=atomic ./...
	@$(call echo_success)


install-env: ## Install required components to develop the project.
	@$(call echo_start)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v${GOLANGCI_VER}
	echo golangci-lint --version
	go install -v github.com/golang/mock/mockgen@latest
	@$(call echo_success)


lint: ## Run linter (https://golangci-lint.run/usage/quick-start/).
	@$(call echo_start)
	golangci-lint run --timeout 3m0s --verbose ./...
	@$(call echo_success)


define gen_mock
	mockgen -source=$(1).go -destination=./mock/$(1).go $(2)
endef
define gen_mock_aux
	mockgen -source=$(1).go -destination=./mock/$(1).go -aux_files=$(3) $(2)
endef
define gen_mock_ext
	-cd ./mock/ && mkdir $(3)
	mockgen $(1) $(2) > ./mock/$(3)/$(3).go 
endef

mock: ## Generate mock interfaces for unit-tests.
	@$(call echo_start)
# "go list ... " in the next run required as a workaround for error - first start mockgen fails with errot at "go list ...":
	-go list -e -compiled=true -test=true ./*
#	$(call gen_mock,service,Service)
	$(call gen_mock_aux,service,Service,${CODE_REPO}=no_copy.go)
	$(call gen_mock_aux,log,Log,${CODE_REPO}=no_copy.go)
# "go list ... " in the next run required as a workaround for error - first start mockgen fails with errot at "go list ...":
	-go list -e -compiled=true -test=true ./ddb/install/*
	$(call gen_mock,ddb/install/db,DB)
	@$(call echo_success)
