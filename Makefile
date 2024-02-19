.PHONY: all build run test clean fmt pre-commit help

TARGET = micro_gateway
ifeq ($(OS),Windows_NT)
TARGET := $(TARGET).exe
endif
TARGET_BIN = $(basename $(TARGET))

ifeq (n$(CGO_ENABLED),n)
CGO_ENABLED := 0
endif

BUILD_ROOT = bin
RELEASE_ROOT = release
RELEASE_FILES = LICENSE README.md run.sh shutdown.sh
RELEASE_LINUX_AMD64 = $(RELEASE_ROOT)/linux-amd64/$(TARGET)
RELEASE_LINUX_AMD64_BIN = $(RELEASE_ROOT)/linux-amd64/$(TARGET)/bin 

RELEASE_DARWIN_AMD64 = $(RELEASE_ROOT)/darwin-amd64/$(TARGET)
RELEASE_DARWIN_AMD64_BIN = $(RELEASE_ROOT)/darwin-amd64/$(TARGET)/bin 

RELEASE_DARWIN_ARM64 = $(RELEASE_ROOT)/darwin-arm64/$(TARGET)
RELEASE_DARWIN_ARM64_BIN = $(RELEASE_ROOT)/darwin-arm64/$(TARGET)/bin 

RELEASE_WINDOWS_AMD64 = $(RELEASE_ROOT)/windows-amd64/$(TARGET)
RELEASE_WINDOWS_AMD64_BIN = $(RELEASE_ROOT)/windows-amd64/$(TARGET)/bin 

BUILD_VERSION := $(shell git describe --tags --always | cut -f1 -f2 -d "-")
BUILD_DATE := $(shell date +'%Y-%m-%d %H:%M:%S')
SHA_SHORT := $(shell git rev-parse --short HEAD)

TAGS = ""
MOD_NAME = github.com/hugokung/micro_gateway
LDFLAGS = -X "${MOD_NAME}/pkg/version.version=${BUILD_VERSION}" \
          -X "${MOD_NAME}/pkg/version.buildDate=${BUILD_DATE}" \
          -X "${MOD_NAME}/pkg/version.commitID=${SHA_SHORT}" -w -s

all: fmt build

build:
	# @go mod download
	@echo Build micro_gateway
	@go build -pgo=auto -trimpath -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $(RELEASE_ROOT)/$(TARGET)

.PHONY: build_dev
build_dev:
	@echo Build micro_gateway
	@go build -pgo=auto -tags '$(TAGS)' -o $(BUILD_ROOT)/$(TARGET)

run:
	@go run -pgo=auto -trimpath -gcflags "all=-N -l" -tags '$(TAGS)' -ldflags '$(LDFLAGS)' .

.PHONY: run_dev
run_dev:
	@go run -pgo=auto -tags '$(TAGS)' .	

.PHONY: release
release:
	@echo Package micro_gateway
	@cp -rf $(RELEASE_FILES) $(RELEASE_LINUX_AMD64) && cp -rf conf $(RELEASE_LINUX_AMD64)/ 
	@cp -rf $(RELEASE_FILES) $(RELEASE_DARWIN_AMD64) && cp -rf conf $(RELEASE_DARWIN_AMD64)/
	@cp -rf $(RELEASE_FILES) $(RELEASE_DARWIN_ARM64) && cp -rf conf $(RELEASE_DARWIN_ARM64)/
	# @cp -rf $(RELEASE_FILES) $(RELEASE_WINDOWS_AMD64) && cp -rf conf $(RELEASE_WINDOWS_AMD64)/
	@cd $(RELEASE_LINUX_AMD64)/.. && rm -f *.zip && zip -r $(TARGET)-linux_amd64.zip $(TARGET) && cd -
	@cd $(RELEASE_DARWIN_AMD64)/.. && rm -f *.zip && zip -r $(TARGET)-darwin_amd64.zip $(TARGET) && cd -
	@cd $(RELEASE_DARWIN_ARM64)/.. && rm -f *.zip && zip -r $(TARGET)-darwin_arm64.zip $(TARGET) && cd -
	# @cd $(RELEASE_WINDOWS_AMD64)/.. && rm -f *.zip && zip -r $(TARGET)-windows_amd64.zip $(TARGET) && cd -

.PHONY: linux-amd64
linux-amd64:
	@echo Build micro_gateway [linux-amd64] CGO_ENABLED=$(CGO_ENABLED)
	@mkdir -p $(RELEASE_LINUX_AMD64)/bin
	# @mkdir -p $(RELEASE_LINUX_AMD64_BIN)
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build -pgo=auto -trimpath -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $(RELEASE_LINUX_AMD64)/bin/$(TARGET_BIN)

.PHONY: darwin-amd64
darwin-amd64:
	@echo Build micro_gateway [darwin-amd64] CGO_ENABLED=$(CGO_ENABLED)
	@mkdir -p $(RELEASE_DARWIN_AMD64)/bin
	# @mkdir -p $(RELEASE_DARWIN_AMD64_BIN)
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=amd64 go build -pgo=auto -trimpath  -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $(RELEASE_DARWIN_AMD64)/bin/$(TARGET_BIN)

.PHONY: darwin-arm64
darwin-arm64:
	@echo Build micro_gateway [darwin-arm64] CGO_ENABLED=$(CGO_ENABLED)
	@mkdir -p $(RELEASE_DARWIN_ARM64)/bin
	# @mkdir -p $(RELEASE_DARWIN_ARM64_BIN)
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=arm64 go build -pgo=auto -trimpath -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $(RELEASE_DARWIN_ARM64)/bin/$(TARGET_BIN)

# .PHONY: windows-x64
# windows-x64:
# 	@echo Build mirco_gateway [windows-x64] CGO_ENABLED=$(CGO_ENABLED)
# 	@mkdir -p $(RELEASE_WINDOWS_AMD64)
# 	@mkdir $(RELEASE_WINDOWS_AMD64_BIN)
# 	@CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=amd64 go build -pgo=auto -trimpath  -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $(RELEASE_WINDOWS_AMD64_BIN)/$(TARGET_BIN).exe

clean:
	@go clean
	@find ./release -type f -exec rm -r {} +

fmt:
	@echo Formatting...
	@go fmt ./internal/...
	@go fmt ./pkg/...
	@go vet -composites=false ./internal/...
	@go vet -composites=false ./pkg/...

pre-commit: fmt
	@go mod tidy

help:
	@echo "make: make"
	@echo "make run: start api server"
	@echo "make build: build executable"
	@echo "make release: build release executables"