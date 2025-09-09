module github.com/appnetorg/online-boutique-arpc

go 1.23.9

replace github.com/appnetorg/online-boutique-arpc/services => ./services

replace github.com/appnetorg/online-boutique-arpc/proto => ./proto

require (
	github.com/appnet-org/arpc v0.0.0-20250908214633-2fd63b0a3e1a
	github.com/appnetorg/online-boutique-arpc/proto v0.0.0-00010101000000-000000000000
	github.com/go-playground/validator/v10 v10.24.0
	github.com/google/uuid v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/redis/go-redis/v9 v9.7.0
	google.golang.org/grpc v1.75.0
	google.golang.org/protobuf v1.36.6
)

require (
	capnproto.org/go/capnp/v3 v3.1.0-alpha.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/colega/zeropool v0.0.0-20230505084239-6fb4a4f75381 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
)
