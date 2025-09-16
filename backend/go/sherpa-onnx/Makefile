CURRENT_DIR=$(abspath ./)
GOCMD=go

ONNX_VERSION?=1.17.1
ONNX_ARCH?=x64
ONNX_OS?=linux

# Detect if we are running on arm64
ifneq (,$(findstring aarch64,$(shell uname -m)))
	ONNX_ARCH=aarch64
endif
ifneq (,$(findstring armv7l,$(shell uname -m)))
	ONNX_ARCH=arm
endif

ifeq ($(OS),Darwin)
	ONNX_OS=osx
	ifneq (,$(findstring aarch64,$(shell uname -m)))
		ONNX_ARCH=arm64
	else ifneq (,$(findstring arm64,$(shell uname -m)))
		ONNX_ARCH=arm64
	else
		ONNX_ARCH=x86_64
	endif
endif

sources/onnxruntime:
	mkdir -p sources/onnxruntime
	curl -L https://github.com/microsoft/onnxruntime/releases/download/v$(ONNX_VERSION)/onnxruntime-$(ONNX_OS)-$(ONNX_ARCH)-$(ONNX_VERSION).tgz -o sources/onnxruntime/onnxruntime-$(ONNX_OS)-$(ONNX_ARCH)-$(ONNX_VERSION).tgz
	cd sources/onnxruntime && tar -xvf onnxruntime-$(ONNX_OS)-$(ONNX_ARCH)-$(ONNX_VERSION).tgz && rm onnxruntime-$(ONNX_OS)-$(ONNX_ARCH)-$(ONNX_VERSION).tgz
	cd sources/onnxruntime && mv onnxruntime-$(ONNX_OS)-$(ONNX_ARCH)-$(ONNX_VERSION)/* ./

backend-assets/lib/libonnxruntime.so.1: sources/onnxruntime
	mkdir -p backend-assets/lib
	cp -rfLv sources/onnxruntime/lib/* backend-assets/lib/
ifeq ($(OS),Darwin)
	mv backend-assets/lib/libonnxruntime.$(ONNX_VERSION).dylib backend-assets/lib/libonnxruntime.dylib
else
	mv backend-assets/lib/libonnxruntime.so.$(ONNX_VERSION) backend-assets/lib/libonnxruntime.so.1
endif

# Copy sherpa-onnx libraries from Go module cache
backend-assets/lib/libsherpa-onnx-c-api.so:
	@echo "Copying sherpa-onnx libraries from Go module cache"
	mkdir -p backend-assets/lib
	# Find the Go module cache path for sherpa-onnx-go-linux
	if [ -z "$$GOPATH" ]; then \
		GO_MOD_CACHE="$$HOME/go/pkg/mod"; \
	else \
		GO_MOD_CACHE="$$GOPATH/pkg/mod"; \
	fi; \
	SHERPA_LIB_DIR=$$(find "$$GO_MOD_CACHE"/github.com/k2-fsa/sherpa-onnx-go-linux* -name "libsherpa-onnx-c-api.so" -exec dirname {} \; 2>/dev/null | head -n 1); \
	if [ -n "$$SHERPA_LIB_DIR" ]; then \
		echo "Found sherpa-onnx libraries at: $$SHERPA_LIB_DIR"; \
		cp -f "$$SHERPA_LIB_DIR"/* backend-assets/lib/; \
		chmod 644 backend-assets/lib/*; \
	else \
		echo "Warning: sherpa-onnx libraries not found in Go module cache"; \
	fi

sherpa-onnx: backend-assets/lib/libonnxruntime.so.1 backend-assets/lib/libsherpa-onnx-c-api.so
	CGO_LDFLAGS="$(CGO_LDFLAGS)" CPATH="$(CPATH):$(CURRENT_DIR)/sources/onnxruntime/include/" LIBRARY_PATH=$(CURRENT_DIR)/backend-assets/lib \
	$(GOCMD) build -ldflags "$(LD_FLAGS)" -tags "$(GO_TAGS)" -o sherpa-onnx ./

clean:
	rm -rf sources/
	rm -rf backend-assets/

package:
	bash package.sh

install-deps:
	go mod tidy

test:
	go test -v ./...

build: install-deps sherpa-onnx package

.PHONY: clean build package test install-deps
