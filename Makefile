.PHONY: gen gen-docker build up down logs

PROTO_FILES := proto/uppercase.proto proto/count.proto proto/gateway.proto

PROTOC_FLAGS := \
	--go_out=. \
	--go-grpc_out=. \
	--go_opt=module=github.com/aberyotaro/grpc-sample \
	--go-grpc_opt=module=github.com/aberyotaro/grpc-sample \
	-I ./proto

# ローカルの protoc で生成する
gen:
	@which protoc > /dev/null 2>&1 || \
		(echo "Error: protoc not found. Install it or run 'make gen-docker' instead." && exit 1)
	protoc $(PROTOC_FLAGS) $(PROTO_FILES)

# Docker を使って生成する (protoc のローカルインストール不要)
gen-docker:
	docker run --rm \
		-v $(PWD):/workspace \
		-w /workspace \
		golang:1.24-bookworm \
		bash -c "\
			apt-get update -q && \
			apt-get install -y -q --no-install-recommends protobuf-compiler && \
			go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
			go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
			protoc $(PROTOC_FLAGS) $(PROTO_FILES)"

build:
	docker compose build

up:
	docker compose up --build

down:
	docker compose down

logs:
	docker compose logs -f
