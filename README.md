# gRPC Sample Project

gRPCで構成したマイクロサービスのサンプル

## アーキテクチャ

```
User (HTTP)
    │
    │ GET /process?text=hello
    ▼
┌─────────────────┐
│  client :8080   │  HTTPサーバー。ユーザーからリクエストを受け取りgatewayへ転送する
└────────┬────────┘
         │ gRPC (GatewayService.Process)
         ▼
┌─────────────────┐
│ gateway :50051  │  API Gateway。リクエストを各バックエンドサービスへルーティングする
└────────┬────────┘
         │
     ┌───┴─────┐
     │         │
     ▼         ▼
┌─────────┐ ┌────────┐
│uppercase│ │ count  │
│ :50052  │ │ :50053 │
└─────────┘ └────────┘
文字列を      文字数を
大文字化      返す
```

## サービス一覧

| サービス  | ポート | 役割 |
|-----------|--------|------|
| client    | 8080   | ユーザー向け HTTP サーバー |
| gateway   | 50051  | gRPC API Gateway |
| uppercase | 50052  | 文字列を大文字に変換する gRPC サービス |
| count     | 50053  | 文字数を返す gRPC サービス |

## ディレクトリ構成

```
.
├── proto/                  # Protocol Buffer 定義
│   ├── gateway.proto
│   ├── uppercase.proto
│   └── count.proto
├── services/               # 各サービスの実装
│   ├── client/main.go
│   ├── gateway/main.go
│   ├── uppercase/main.go
│   └── count/main.go
├── Dockerfile.client
├── Dockerfile.gateway
├── Dockerfile.uppercase
├── Dockerfile.count
├── docker-compose.yml
├── go.mod
└── Makefile
```

> **Note**
> `gen/` ディレクトリは protoc が自動生成するため、リポジトリには含まれていません。
> Docker ビルド時に各 Dockerfile 内で生成されます。

## 起動方法

```bash
docker compose up --build
```

### 動作確認

```bash
curl "http://localhost:8080/process?text=hello"
```

```json
{"count":5,"original":"hello","uppercase":"HELLO"}
```

### 停止

```bash
docker compose down
```

---

## 新しいサービスを追加する手順

ここでは例として **逆順にする** `reverse` サービスを追加する手順を説明します。

### 1. Proto ファイルを追加する

`proto/reverse.proto` を作成します。

```proto
syntax = "proto3";

package reverse;

option go_package = "github.com/aberyotaro/grpc-sample/gen/reverse";

service ReverseService {
  rpc Reverse(ReverseRequest) returns (ReverseResponse);
}

message ReverseRequest {
  string text = 1;
}

message ReverseResponse {
  string text = 1;
}
```

### 2. サービスの実装を追加する

`services/reverse/main.go` を作成します。

```go
package main

import (
    "context"
    "log"
    "net"

    pb "github.com/aberyotaro/grpc-sample/gen/reverse"
    "google.golang.org/grpc"
)

type server struct {
    pb.UnimplementedReverseServiceServer
}

func (s *server) Reverse(_ context.Context, req *pb.ReverseRequest) (*pb.ReverseResponse, error) {
    runes := []rune(req.Text)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return &pb.ReverseResponse{Text: string(runes)}, nil
}

func main() {
    lis, err := net.Listen("tcp", ":50054")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    s := grpc.NewServer()
    pb.RegisterReverseServiceServer(s, &server{})
    log.Println("reverse service listening on :50054")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
```

### 3. Dockerfile を追加する

`Dockerfile.reverse` を作成します。
既存の `Dockerfile.uppercase` をベースに、サービス名とポートを変えるだけです。

```dockerfile
# Stage 1: proto code generation
FROM golang:1.24-bookworm AS proto-builder

RUN apt-get update && apt-get install -y --no-install-recommends protobuf-compiler \
    && rm -rf /var/lib/apt/lists/*

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

COPY proto/ /proto/

RUN protoc \
    --go_out=/ \
    --go-grpc_out=/ \
    --go_opt=module=github.com/aberyotaro/grpc-sample \
    --go-grpc_opt=module=github.com/aberyotaro/grpc-sample \
    -I /proto \
    /proto/reverse.proto

# Stage 2: build
FROM golang:1.24-bookworm AS builder

WORKDIR /app

COPY go.mod ./
COPY --from=proto-builder /gen ./gen/
COPY services/reverse/ ./services/reverse/

RUN go mod tidy && go build -o /server ./services/reverse/

# Stage 3: runtime
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /server /server

EXPOSE 50054

CMD ["/server"]
```

### 4. gateway を更新する

`gateway` が新しいサービスを呼び出せるように変更します。

**`proto/gateway.proto`** — レスポンスにフィールドを追加

```proto
message ProcessResponse {
  string uppercase = 1;
  int32  count     = 2;
  string reverse   = 3;  // 追加
}
```

**`services/gateway/main.go`** — reverse サービスに接続してレスポンスに含める

```go
// 接続を追加
reverseConn, err := grpc.NewClient("reverse:50054",
    grpc.WithTransportCredentials(insecure.NewCredentials()))

// server struct に追加
type server struct {
    ...
    reverseClient reversepb.ReverseServiceClient
}

// Process メソッドで呼び出す
reverseRes, err := s.reverseClient.Reverse(ctx, &reversepb.ReverseRequest{Text: req.Text})
```

**`Dockerfile.gateway`** — reverse.proto も一緒に生成するよう追記

```dockerfile
RUN protoc ... /proto/gateway.proto /proto/uppercase.proto /proto/count.proto /proto/reverse.proto
```

### 5. docker-compose.yml にサービスを追加する

```yaml
  reverse:
    build:
      context: .
      dockerfile: Dockerfile.reverse
    ports:
      - "50054:50054"
```

gateway の `depends_on` にも追加します。

```yaml
  gateway:
    depends_on:
      - uppercase
      - count
      - reverse  # 追加
```

### 6. 起動して確認

```bash
docker compose up --build
curl "http://localhost:8080/process?text=hello"
# {"count":5,"original":"hello","reverse":"olleh","uppercase":"HELLO"}
```

---

## ローカルで proto を再生成する (オプション)

### Docker を使う方法（推奨・インストール不要）

```bash
make gen-docker
```

### protoc をローカルにインストールする方法

```bash
# protoc-gen-go / protoc-gen-go-grpc をインストール
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 生成
make gen
```
