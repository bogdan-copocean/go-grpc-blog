genproto:
	protoc proto/blog.proto \
	--js_out=import_style=commonjs,binary:./ui/src/ \
	--grpc-web_out=import_style=commonjs,mode=grpcwebtext:./ui/src/ \
	--go_out=. \
	--go_opt=paths=source_relative \
	--go-grpc_out=. \
	--go-grpc_opt=paths=source_relative 
