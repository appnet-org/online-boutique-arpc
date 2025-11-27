package services

import (
	"fmt"
	"log"
	"os"

	"github.com/appnet-org/arpc/pkg/custom/reliable"
	"github.com/appnet-org/arpc/pkg/logging"
	"github.com/appnet-org/arpc/pkg/packet"
	"github.com/appnet-org/arpc/pkg/rpc"
	"github.com/appnet-org/arpc/pkg/rpc/element"
	"github.com/appnet-org/arpc/pkg/serializer"
	"github.com/appnet-org/arpc/pkg/transport"
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

// mustConnARPC creates an aRPC client with tracing and reliable delivery
func mustConnARPC(client **rpc.Client, addr string) {
	log.Printf("Attempting to connect to aRPC server at: %s", addr)

	serializer := &serializer.SymphonySerializer{}
	clientElements := []element.RPCElement{tracing.NewClientTracingElement()}

	var err error
	// Use 0.0.0.0:0 to explicitly bind to IPv4 instead of :0 which defaults to IPv6
	c, err := rpc.NewClientWithLocalAddr(serializer, addr, "0.0.0.0:0", clientElements)
	if err != nil {
		panic(errors.Wrapf(err, "arpc: failed to connect %s", addr))
	}

	// Get UDP transport from the client
	udpTransport := c.Transport()

	// Register packet type for reliable delivery
	ackPacketType, err := udpTransport.RegisterPacketType(
		reliable.AckPacketName,
		&reliable.ACKPacketCodec{})
	if err != nil {
		panic(errors.Wrap(err, "failed to register ACK packet type"))
	}

	// Create handler for reliable delivery
	reliableHandler := reliable.NewReliableClientHandler(
		udpTransport,
		udpTransport.GetTimerManager(),
	)

	// Get existing handler chains for REQUEST packets (OnSend)
	requestChain, exists := udpTransport.GetHandlerRegistry().GetHandlerChain(
		packet.PacketTypeRequest.TypeID,
		transport.RoleClient,
	)
	if !exists {
		panic("failed to get REQUEST handler chain")
	}
	requestChain.AddHandler(reliableHandler)

	// Get existing handler chains for RESPONSE packets (OnReceive)
	responseChain, exists := udpTransport.GetHandlerRegistry().GetHandlerChain(
		packet.PacketTypeResponse.TypeID,
		transport.RoleClient,
	)
	if !exists {
		panic("failed to get RESPONSE handler chain")
	}
	responseChain.AddHandler(reliableHandler)

	// Register dedicated handler chain for ACK packets
	ackChain := transport.NewHandlerChain("ClientACKHandlerChain", reliableHandler)
	udpTransport.RegisterHandlerChain(ackPacketType.TypeID, ackChain, transport.RoleClient)

	*client = c
}

// setupServerReliableCC configures reliable delivery for an aRPC server
// This function must be called after rpc.NewServer() but before server.Start()
// Returns a cleanup function that should be deferred
func setupServerReliableCC(server *rpc.Server) func() {
	// Get the UDP transport from the server
	udpTransport := server.GetTransport()

	// Register ACK packet type for reliable delivery
	ackPacketType, err := udpTransport.RegisterPacketType(reliable.AckPacketName, &reliable.ACKPacketCodec{})
	if err != nil {
		log.Fatalf("Failed to register ACK packet type: %v", err)
	}

	// Create reliable server handler
	reliableHandler := reliable.NewReliableServerHandler(
		udpTransport,
		udpTransport.GetTimerManager(),
	)

	// Get handler chains for REQUEST packets (OnReceive)
	requestChain, exists := udpTransport.GetHandlerRegistry().GetHandlerChain(
		packet.PacketTypeRequest.TypeID,
		transport.RoleServer,
	)
	if !exists {
		log.Fatal("Failed to get REQUEST handler chain")
	}
	requestChain.AddHandler(reliableHandler)

	// Get handler chains for RESPONSE packets (OnSend)
	responseChain, exists := udpTransport.GetHandlerRegistry().GetHandlerChain(
		packet.PacketTypeResponse.TypeID,
		transport.RoleServer,
	)
	if !exists {
		log.Fatal("Failed to get RESPONSE handler chain")
	}
	responseChain.AddHandler(reliableHandler)

	// Register handler chain for ACK packets
	ackChain := transport.NewHandlerChain("ServerACKHandlerChain", reliableHandler)
	udpTransport.RegisterHandlerChain(ackPacketType.TypeID, ackChain, transport.RoleServer)

	// Return cleanup function
	return func() {
		reliableHandler.Cleanup()
	}
}
