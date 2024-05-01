# default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# The name of Terraform custom provider.
CUSTOM_PROVIDER_NAME ?= terraform-provider-domain-management
# The url of Terraform provider.
CUSTOM_PROVIDER_URL ?= example.local/myklst/domain-management

UNAME := $(shell uname)

.PHONY: install-local-domain-management
install-local-domain-management: linux_amd64

linux_amd64:
ifneq ($(UNAME), Linux)
	$(info 'skip linux_amd64')
else
	export PROVIDER_LOCAL_PATH='$(CUSTOM_PROVIDER_URL)'
	GOOS=linux GOARCH=amd64 go build -o $(CUSTOM_PROVIDER_NAME) .
	HOME_DIR="$$(ls -d ~)"; \
	mkdir -p  $$HOME_DIR/.terraform.d/plugins/$(CUSTOM_PROVIDER_URL)/0.1.0/linux_amd64/; \
	cp ./$(CUSTOM_PROVIDER_NAME) $$HOME_DIR/.terraform.d/plugins/$(CUSTOM_PROVIDER_URL)/0.1.0/linux_amd64/$(CUSTOM_PROVIDER_NAME)
	unset PROVIDER_LOCAL_PATH
endif
