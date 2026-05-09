NAME     = ecms
VERSION ?= 0.1.0

GO             ?= go
GOARCH         ?= $(shell $(GO) env GOARCH)
GO_MODULE      ?= github.com/kweaver-ai/proton/$(NAME)
GO_FILES       := $(wildcard **/*.go)
GO_BUILD_FLAGS ?= -ldflags "$(LD_FLAGS)"
LD_FLAGS       ?= -s -w
CGO_ENABLED    ?= 0

RPM_ARCH_amd64 = x86_64
RPM_ARCH_arm64 = aarch_64

RPM_VERSION ?= $(VERSION)
RPM_RELEASE ?= 1.proton
RPM_ARCH    ?= $(RPM_ARCH_$(GOARCH))
RPM_FILE    ?= $(NAME)-$(RPM_VERSION)-$(RPM_RELEASE).$(RPM_ARCH).rpm


build: bin/$(NAME)

bin/$(NAME): $(GO_FILES) go.mod go.sum
	CGO_ENABLED=$(CGO_ENABLED) GOARCH=$(GOARCH) $(GO) build -o $(@) $(GO_BUILD_FLAGS) "$(GO_MODULE)/cmd/$(NAME)"

release: $(RPM_FILE)

$(RPM_FILE): bin/$(NAME)
	VERSION=$(RPM_BUILD) RELEASE=$(RPM_RELEASE) ARCH=$(RPM_ARCH) nfpm package --packager=rpm --target=$(RPM_FILE)

.PHONY: clean
clean:
	@rm -fv bin/$(NAME) $(RPM_FILE)
