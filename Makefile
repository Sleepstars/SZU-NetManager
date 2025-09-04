# Default environment (override with: make run-backend NM_SSH_HOST=192.168.1.1)
NM_LISTEN            ?= :8080
NM_DB                ?= szu-netmanager.db
NM_SSH_HOST          ?= 127.0.0.1
NM_SSH_PORT          ?= 22
NM_SSH_USER          ?= root
NM_SSH_KEY           ?= $(HOME)/.ssh/id_rsa
NM_MONITOR_INTERVAL  ?= 30
NM_MONITOR_URLS      ?= https://www.baidu.com,https://www.qq.com
NM_SZU_LOGIN         ?= /usr/local/bin/srun-login
NM_WEB_DIR           ?= web/dist

GO                   ?= go
NODE                 ?= npm

BIN_DIR              := bin
BACKEND_BIN          := $(BIN_DIR)/netmanager
BACKEND_PKG          := ./cmd/netmanager
WEB_DIR              := web
WEB_DIST             := $(WEB_DIR)/dist

DOCKER_IMAGE         ?= szu-netmanager
SZU_LOGIN_URL        ?= https://github.com/Sleepstars/SZU-login/releases/latest/download/srun-login-linux-amd64

.PHONY: help init deps build-backend run-backend build-frontend dev-frontend build-all clean docker-build docker-run print-env

help:
	@echo "Targets:"
	@echo "  init             - One-time setup: go mod tidy + npm install"
	@echo "  deps             - Same as init"
	@echo "  build-backend    - Build Go backend to $(BACKEND_BIN)"
	@echo "  run-backend      - Run backend with NM_* env vars"
	@echo "  build-frontend   - Build React app to $(WEB_DIST)"
	@echo "  dev-frontend     - Start Vite dev server"
	@echo "  build-all        - Build frontend then backend"
	@echo "  docker-build     - Build multi-stage Docker image"
	@echo "  docker-run       - Run image with host network and NM_* env"
	@echo "  print-env        - Show effective NM_* variables"
	@echo "  clean            - Remove build artifacts"

init: deps

deps:
	$(GO) mod tidy
	cd $(WEB_DIR) && ($(NODE) ci || $(NODE) i)

$(BACKEND_BIN):
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BACKEND_BIN) $(BACKEND_PKG)

build-backend: $(BACKEND_BIN)

run-backend:
	NM_LISTEN=$(NM_LISTEN) \
	NM_DB=$(NM_DB) \
	NM_SSH_HOST=$(NM_SSH_HOST) \
	NM_SSH_PORT=$(NM_SSH_PORT) \
	NM_SSH_USER=$(NM_SSH_USER) \
	NM_SSH_KEY=$(NM_SSH_KEY) \
	NM_MONITOR_INTERVAL=$(NM_MONITOR_INTERVAL) \
	NM_MONITOR_URLS=$(NM_MONITOR_URLS) \
	NM_SZU_LOGIN=$(NM_SZU_LOGIN) \
	NM_WEB_DIR=$(NM_WEB_DIR) \
	$(GO) run $(BACKEND_PKG)

build-frontend:
	cd $(WEB_DIR) && ($(NODE) ci || $(NODE) i)
	cd $(WEB_DIR) && $(NODE) run build

dev-frontend:
	cd $(WEB_DIR) && ($(NODE) ci || $(NODE) i)
	cd $(WEB_DIR) && $(NODE) run dev

build-all: build-frontend build-backend

docker-build:
	docker build --build-arg SZU_LOGIN_URL=$(SZU_LOGIN_URL) -t $(DOCKER_IMAGE) .

docker-run:
	docker run --rm --network host \
	  -e NM_LISTEN=$(NM_LISTEN) \
	  -e NM_DB=$(NM_DB) \
	  -e NM_SSH_HOST=$(NM_SSH_HOST) \
	  -e NM_SSH_PORT=$(NM_SSH_PORT) \
	  -e NM_SSH_USER=$(NM_SSH_USER) \
	  -e NM_SSH_KEY=$(NM_SSH_KEY) \
	  -e NM_MONITOR_INTERVAL=$(NM_MONITOR_INTERVAL) \
	  -e NM_MONITOR_URLS=$(NM_MONITOR_URLS) \
	  -e NM_SZU_LOGIN=$(NM_SZU_LOGIN) \
	  -e NM_WEB_DIR=$(NM_WEB_DIR) \
	  $(DOCKER_IMAGE)

print-env:
	@echo NM_LISTEN=$(NM_LISTEN)
	@echo NM_DB=$(NM_DB)
	@echo NM_SSH_HOST=$(NM_SSH_HOST)
	@echo NM_SSH_PORT=$(NM_SSH_PORT)
	@echo NM_SSH_USER=$(NM_SSH_USER)
	@echo NM_SSH_KEY=$(NM_SSH_KEY)
	@echo NM_MONITOR_INTERVAL=$(NM_MONITOR_INTERVAL)
	@echo NM_MONITOR_URLS=$(NM_MONITOR_URLS)
	@echo NM_SZU_LOGIN=$(NM_SZU_LOGIN)
	@echo NM_WEB_DIR=$(NM_WEB_DIR)
	@echo DOCKER_IMAGE=$(DOCKER_IMAGE)
	@echo SZU_LOGIN_URL=$(SZU_LOGIN_URL)

clean:
	rm -rf $(BIN_DIR) $(WEB_DIST)

