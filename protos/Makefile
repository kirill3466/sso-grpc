gen_grpc_sso:
	protoc -I proto -I third_party proto/sso/*.proto \
		--go_out=./gen/go/ --go_opt=paths=source_relative \
		--go-grpc_out=./gen/go/ --go-grpc_opt=paths=source_relative

gen_grpc_rtdb:
	protoc -I proto -I third_party proto/rtdb/*.proto \
		--go_out=./gen/go/ --go_opt=paths=source_relative \
		--go-grpc_out=./gen/go/ --go-grpc_opt=paths=source_relative

tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
