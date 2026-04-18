# Always override shell GOPROXY: a global goproxy.io often returns 502 for modernc.org/*.
# To pick another mirror: make dev GO_MOD_PROXY=https://goproxy.cn,direct
GO_MOD_PROXY ?= https://proxy.golang.org,direct

.PHONY: dev tidy build

dev:
	GOPROXY=$(GO_MOD_PROXY) wails dev

tidy:
	GOPROXY=$(GO_MOD_PROXY) go mod tidy

build:
	GOPROXY=$(GO_MOD_PROXY) wails build
