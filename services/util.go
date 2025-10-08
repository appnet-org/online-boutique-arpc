package services

import (
	"fmt"
	"log"
	"os"

	"github.com/appnet-org/arpc/pkg/rpc"
	"github.com/appnet-org/arpc/pkg/rpc/element"
	"github.com/appnet-org/arpc/pkg/serializer"
	"github.com/appnetorg/online-boutique-arpc/services/tracing"
	"github.com/pkg/errors"
)

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

	serializer := &serializer.SymphonySerializer{}
	clientElements := []element.RPCElement{tracing.NewClientTracingElement()}

	var err error
	*client, err = rpc.NewClient(serializer, addr, clientElements)
	if err != nil {
		panic(errors.Wrapf(err, "arpc: failed to connect %s", addr))
	}
}
