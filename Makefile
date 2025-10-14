secret.pb.go secret_grpc.pb.go: secret.proto
	protoc --proto_path=. \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		./secret.proto
