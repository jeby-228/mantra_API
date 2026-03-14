# 會員 API 專案的 Justfile
default:
    @just --list

# 初始化開發環境
setup:
    @echo "安裝 golangci-lint..."
    @curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin
    @echo "安裝 swag..."
    @SWAG_BIN="${GOBIN:-$(go env GOPATH)/bin}/swag"; \
    if [ ! -x "$SWAG_BIN" ]; then \
        go install github.com/swaggo/swag/cmd/swag@latest; \
    fi; \
    "$SWAG_BIN" --version
    @echo "安裝相依套件..."
    go mod download
    @echo "設定完成！"

# 本地運行
run:
    go run main.go

# 編譯
build:
    go build

# 測試
test:
    go test -v ./...

# 格式化
fmt:
    go fmt ./...

# 程式碼檢查
lint:
    ./bin/golangci-lint run ./...

# 程式碼檢查並自動修復
lint-fix:
    ./bin/golangci-lint run --fix ./...

# 整理依賴
tidy:
    go mod tidy

# 測試覆蓋率
test-coverage:
    go test -cover ./...

# 測試覆蓋率產生報告
test-cov:
    @echo "產生測試覆蓋率報告..."
    @go test -v -coverprofile=coverage.out ./auth ./config ./services
    @if [ -f coverage.out ]; then \
        go tool cover -html=coverage.out -o coverage.html && \
        echo "測試覆蓋率報告已產生：coverage.html"; \
    else \
        echo "沒有可用的覆蓋率資料"; \
    fi


# 生成 GraphQL
graphql:
    go get github.com/99designs/gqlgen@v0.17.85
    go run github.com/99designs/gqlgen generate

# 生成 swagger
swag:
    @SWAG_BIN="${GOBIN:-$(go env GOPATH)/bin}/swag"; \
    [ -x "$SWAG_BIN" ] || SWAG_BIN="$(command -v swag 2>/dev/null || true)"; \
    [ -n "$SWAG_BIN" ] || { echo "swag 尚未安裝，請先執行 just setup"; exit 1; }; \
    "$SWAG_BIN" init

# 清理
clean:
    rm -rf bin/ coverage.out coverage.html

# 一鍵檢查
auto_check:
    just tidy
    just fmt
    just lint
    just test
    just build
    just graphql
    just test-cov
