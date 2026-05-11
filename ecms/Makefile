NAME     = ecms
VERSION ?= 0.1.0

GO             ?= go
GOARCH         ?= $(shell $(GO) env GOARCH)
GO_MODULE      ?= github.com/kweaver-ai/proton/ecms
GO_FILES       := $(shell find * -type f -name "*.go")
GO_BUILD_FLAGS ?= -ldflags "$(LD_FLAGS)"
LD_FLAGS       ?= -s -w
CGO_ENABLED    ?= 0

RPM_ARCH_amd64 = x86_64
RPM_ARCH_arm64 = aarch_64

RPM_VERSION ?= $(VERSION)
RPM_RELEASE ?= 1.proton
RPM_ARCH    ?= $(RPM_ARCH_$(GOARCH))
RPM_SCRIPTS ?= $(shell find scripts -type f)


build: bin/$(NAME)

bin/$(NAME): $(GO_FILES) go.mod go.sum
	CGO_ENABLED=$(CGO_ENABLED) GOARCH=$(GOARCH) $(GO) build -o $(@) $(GO_BUILD_FLAGS) "$(GO_MODULE)/cmd/$(NAME)"

release: $(RPM_FILE)

$(RPM_FILE): $(BIN_FILE) $(NFPM_CONFIG)
	VERSION=$(RPM_BUILD) RELEASE=$(RPM_RELEASE) ARCH=$(RPM_ARCH) nfpm package --config=$(NFPM_CONFIG) --packager=rpm

.PHONY: clean
clean:
	@rm -fv bin/$(NAME) *.rpm
