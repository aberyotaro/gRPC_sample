# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# 全サービスをビルドして起動
make up  # docker compose up --build

# 起動中のサービスを停止
make down

# ログを確認
make logs

# proto から Go コードを生成（protoc 不要）
make gen-docker

# proto から Go コードを生成（ローカル protoc が必要）
make gen

# 動作確認
curl "http://localhost:8080/process?text=hello"
```

## アーキテクチャ

```
User (HTTP) → client:8080 → gateway:50051 → uppercase:50052
                                           → count:50053
```

4サービスがそれぞれ独立した Docker コンテナで動作する。

- **client**: HTTP サーバー。ユーザーリクエストを受けて gateway gRPC を呼び出す
- **gateway**: API Gateway。upstream の uppercase・count を並列呼び出しして結果をまとめる
- **uppercase**: 文字列を大文字化する gRPC サービス
- **count**: 文字数（rune 数）を返す gRPC サービス

## Proto と Go コード生成

`gen/` ディレクトリはリポジトリに含まれず、Docker ビルド時に各 Dockerfile 内で生成される。

生成の仕組み（各 Dockerfile の Stage 1）：
```
protoc --go_out=/ --go-grpc_out=/ \
    --go_opt=module=github.com/aberyotaro/grpc-sample \
    ...
```

`--go_out=/` を使う理由: `--go_out=/gen` にすると `go_package` のパスと合わさって `/gen/gen/...` と二重になるため、`/` を起点にすることで `/gen/{service}/` に正しく出力される。

`go_package` は `github.com/aberyotaro/grpc-sample/gen/{service}` の形式で定義されており、モジュールプレフィックスを除いた `gen/{service}/` が出力パスになる。

## 新しいサービスを追加するときのチェックリスト

1. `proto/{name}.proto` を作成（`go_package = "github.com/aberyotaro/grpc-sample/gen/{name}"`）
2. `services/{name}/main.go` を実装
3. `Dockerfile.{name}` を作成（既存の Dockerfile からコピーしてポートと proto ファイル名を変更）
4. `docker-compose.yml` にサービスを追加
5. `gateway` から呼び出す場合は `proto/gateway.proto` のレスポンスにフィールド追加、`services/gateway/main.go` を更新、`Dockerfile.gateway` の protoc コマンドに新 proto を追加
6. `Makefile` の `PROTO_FILES` に新 proto を追加

## モジュール

`module github.com/aberyotaro/grpc-sample`（Go 1.24）

主要な依存：
- `google.golang.org/grpc v1.79.1`
- `google.golang.org/protobuf v1.36.11`
