.PHONY: all build run clean test lint docker help

# 基本变量
APP_NAME=distance-back
MAIN_FILE=cmd/app/main.go
BUILD_DIR=build
DOCKER_IMAGE=$(APP_NAME)
GIT_COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go 相关命令
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get

# 默认目标
all: lint test build

# 构建应用
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags="-X main.Version=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME)" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "Build complete"

# 运行开发环境
run-dev:
	$(GORUN) $(MAIN_FILE) -env development

# 运行生产环境
run-prod:
	$(GORUN) $(MAIN_FILE) -env production

# 清理构建文件
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	$(GOCLEAN)
	@echo "Clean complete"

# 运行测试
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# 运行指定包的测试
test-package:
	@if [ "$(pkg)" = "" ]; then \
		echo "Please specify a package: make test-package pkg=./internal/service"; \
	else \
		$(GOTEST) -v $(pkg); \
	fi

# 代码格式化和检查
lint:
	@echo "Running linters..."
	golangci-lint run
	@echo "Lint complete"

# 生成 Docker 镜像
docker-build:
	docker build -t $(DOCKER_IMAGE):$(GIT_COMMIT) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		.

# 推送 Docker 镜像
docker-push:
	docker push $(DOCKER_IMAGE):$(GIT_COMMIT)

# 更新依赖
deps:
	$(GOGET) -u ./...
	go mod tidy

# 生成 mock 代码
mock:
	@echo "Generating mocks..."
	mockgen -source=internal/repository/repository.go -destination=test/mock/repository_mock.go -package=mock

# 创建数据库迁移文件
migrate-create:
	@if [ "$(name)" = "" ]; then \
		echo "Please specify a migration name: make migrate-create name=create_users_table"; \
	else \
		migrate create -ext sql -dir scripts/migrations -seq $(name); \
	fi

# 运行数据库迁移
migrate-up:
	migrate -path scripts/migrations -database "mysql://$(DB_USER):$(DB_PASS)@tcp($(DB_HOST):3306)/$(DB_NAME)" up

# 回滚数据库迁移
migrate-down:
	migrate -path scripts/migrations -database "mysql://$(DB_USER):$(DB_PASS)@tcp($(DB_HOST):3306)/$(DB_NAME)" down

# 显示帮助信息
help:
	@echo "Available targets:"
	@echo "  make build          - Build the application"
	@echo "  make run-dev        - Run in development mode"
	@echo "  make run-prod       - Run in production mode"
	@echo "  make clean          - Clean build files"
	@echo "  make test           - Run all tests"
	@echo "  make test-package   - Run tests for a specific package"
	@echo "  make lint           - Run linters"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-push    - Push Docker image"
	@echo "  make deps           - Update dependencies"
	@echo "  make mock           - Generate mock files"
	@echo "  make migrate-create - Create a new migration file"
	@echo "  make migrate-up     - Run database migrations"
	@echo "  make migrate-down   - Rollback database migrations"