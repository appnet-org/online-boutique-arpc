module github.com/appnetorg/online-boutique-arpc

go 1.24.0

replace github.com/appnetorg/online-boutique-arpc/services => ./services

replace github.com/appnetorg/online-boutique-arpc/proto => ./proto

require (
	github.com/appnet-org/arpc-quic v0.0.0-20251130031928-0f2e1f8f8ccb
	github.com/appnetorg/online-boutique-arpc/proto v0.0.0-00010101000000-000000000000
	github.com/go-playground/validator/v10 v10.28.0
	github.com/google/uuid v1.6.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/redis/go-redis/v9 v9.14.0
	github.com/uber/jaeger-client-go v2.30.0+incompatible
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
)

require (
	capnproto.org/go/capnp/v3 v3.1.0-alpha.2 // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/appnet-org/arpc-tcp v0.0.0-20251112223301-306d4c0c7f8c // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/colega/zeropool v0.0.0-20230505084239-6fb4a4f75381 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gabriel-vasile/mimetype v1.4.10 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/google/pprof v0.0.0-20210407192527-94a9f03dee38 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/onsi/ginkgo/v2 v2.9.5 // indirect
	github.com/quic-go/qtls-go1-20 v0.4.1 // indirect
	github.com/quic-go/quic-go v0.40.1 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/mock v0.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.43.0 // indirect
	golang.org/x/exp v0.0.0-20240604190554-fc45aab8b7f8 // indirect
	golang.org/x/mod v0.28.0 // indirect
	golang.org/x/net v0.45.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	golang.org/x/tools v0.37.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251007200510-49b9836ed3ff // indirect
)
