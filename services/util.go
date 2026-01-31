package services

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/appnet-org/arpc-h2/pkg/logging"
	"github.com/appnet-org/arpc-h2/pkg/rpc"
	"github.com/appnet-org/arpc-h2/pkg/rpc/element"
	"github.com/appnet-org/arpc-h2/pkg/serializer"
	"github.com/appnetorg/online-boutique-arpc/services/tracing"
	"github.com/pkg/errors"
)

// getLoggingConfig reads logging configuration from environment variables with defaults
func getLoggingConfig() *logging.Config {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	format := os.Getenv("LOG_FORMAT")
	if format == "" {
		format = "console"
	}

	return &logging.Config{
		Level:  level,
		Format: format,
	}
}

func mustMapEnv(target *string, envKey string) {
	v := os.Getenv(envKey)
	if v == "" {
		panic(fmt.Sprintf("environment variable %q not set", envKey))
	}
	*target = v
}

// mustConnARPC creates an aRPC client with tracing, similar to mustConnGRPC
func mustConnARPC(client **rpc.Client, addr string) {
	log.Printf("Attempting to connect to aRPC server at: %s", addr)

	serializer := &serializer.ProtoSerializer{}
	clientElements := []element.RPCElement{tracing.NewClientTracingElement()}

	var err error
	// Use 0.0.0.0:0 to explicitly bind to IPv4 instead of :0 which defaults to IPv6
	*client, err = rpc.NewClientWithLocalAddr(serializer, addr, "0.0.0.0:0", clientElements...)
	if err != nil {
		panic(errors.Wrapf(err, "arpc: failed to connect %s", addr))
	}

	// Optional streaming mode to avoid per-RPC HTTP request overhead.
	if v := os.Getenv("ARPC_STREAMING_MODE"); v != "" {
		enabled, err := strconv.ParseBool(v)
		if err != nil {
			log.Printf("Invalid ARPC_STREAMING_MODE=%q; expected true/false", v)
		} else if enabled {
			if err := (*client).EnableStreaming(); err != nil {
				panic(errors.Wrap(err, "arpc: failed to enable streaming mode"))
			}
			log.Printf("aRPC streaming mode enabled for %s", addr)
		}
	}

	if err := (*client).EnableStreaming(); err != nil {
		panic(errors.Wrap(err, "arpc: failed to enable streaming mode"))
	}
	log.Printf("aRPC streaming mode enabled for %s", addr)
}
