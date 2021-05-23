GOTEST=go test

FORCE: help

all: FORCE ## This is a library, nothing to build

test: FORCE ## Run tests (todo fix this)
		$(GOTEST) -v ./pkg/...

clean: ## Clean distribution files
	go clean
	rm -f test/certs/*
	rm -f test/logs/*

#deps: ## Build GO dependencies 
#		$(GODEP)

#upgrade: ## Upgrade the dependencies to the latest version. Use with care.
#		go fix


help: ## Show this help
		@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
