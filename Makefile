# Copyright 2021, the SS project owners. All rights reserved.
# Please see the OWNERS and LICENSE files for details.

.PHONY: all \
	install-mock install-mock-deps \
	lint \
	mock \
.DEFAULT_GOAL = all

GO_GET_CMD = go get -v

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
	$(call make_target,install-mock)
	$(call make_target,lint)
	@$(call echo_success)


install-mock: ## Install mock compilator and generate mock.
	@$(call echo_start)
	$(call make_target,install-mock-deps)
	$(call make_target,mock)
	@$(call echo_success)
install-mock-deps: ## Install mock compilator components.
	@$(call echo_start)
	${GO_GET_CMD} github.com/stretchr/testify/assert
	${GO_GET_CMD} github.com/golang/mock/gomock
	${GO_GET_CMD} github.com/golang/mock/mockgen
	@$(call echo_success)

lint: ## Run linter.
	@$(call echo_start)
	golangci-lint run -v ./...
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
	$(call gen_mock,service,Service)
	$(call gen_mock,log,Log)
# "go list ... " in the next run required as a workaround for error - first start mockgen fails with errot at "go list ...":
	-go list -e -compiled=true -test=true ./ddb/install/*
	$(call gen_mock,ddb/install/db,DB)
	@$(call echo_success)
