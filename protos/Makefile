generate grpc:
	protoc -I proto proto/sso/*.proto \
		--go_out=./gen/go/ --go_opt=paths=source_relative \
		--go-grpc_out=./gen/go/ --go-grpc_opt=paths=source_relative

tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
