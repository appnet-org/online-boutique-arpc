package services

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/appnet-org/arpc/pkg/custom/congestion"
	"github.com/appnet-org/arpc/pkg/custom/flowcontrol"
	"github.com/appnet-org/arpc/pkg/custom/reliable"
	"github.com/appnet-org/arpc/pkg/logging"
	"github.com/appnet-org/arpc/pkg/packet"
	"github.com/appnet-org/arpc/pkg/rpc"
	"github.com/appnet-org/arpc/pkg/rpc/element"
	"github.com/appnet-org/arpc/pkg/serializer"
	"github.com/appnet-org/arpc/pkg/transport"
	"github.com/appnetorg/online-boutique-arpc/services/messagelogger"
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

// parseEnvBool parses an environment variable as a boolean
// Supports: "true", "1", "yes", "on" (case-insensitive) -> true
//
//	"false", "0", "no", "off" (case-insensitive) -> false
//
// Returns defaultValue if the variable is not set or invalid
func parseEnvBool(envKey string, defaultValue bool) bool {
	enableStr := os.Getenv(envKey)
	if enableStr == "" {
		return defaultValue
	}

	// Parse the value - support "true", "1", "yes", "on" (case-insensitive)
	enableStr = strings.ToLower(strings.TrimSpace(enableStr))
	if enableStr == "true" || enableStr == "1" || enableStr == "yes" || enableStr == "on" {
		return true
	}
	if enableStr == "false" || enableStr == "0" || enableStr == "no" || enableStr == "off" {
		return false
	}

	// Try parsing as boolean
	enabled, err := strconv.ParseBool(enableStr)
	if err != nil {
		log.Printf("Warning: Invalid value for %s: %q, defaulting to %v", envKey, enableStr, defaultValue)
		return defaultValue
	}
	return enabled
}

// isReliableEnabled checks if reliable delivery feature should be enabled
// It reads the ENABLE_RELIABLE environment variable (defaults to true if not set)
func isReliableEnabled() bool {
	return parseEnvBool("ENABLE_RELIABLE", true)
}

// isCCEnabled checks if congestion control feature should be enabled
// It reads the ENABLE_CC environment variable (defaults to true if not set)
func isCCEnabled() bool {
	return parseEnvBool("ENABLE_CC", true)
}

// isFCEnabled checks if flow control feature should be enabled
// It reads the ENABLE_FC environment variable (defaults to true if not set)
func isFCEnabled() bool {
	return parseEnvBool("ENABLE_FC", true)
}

// isEncryptionEnabled checks if encryption feature should be enabled
// It reads the ENABLE_ENCRYPTION environment variable (defaults to false if not set)
func isEncryptionEnabled() bool {
	return parseEnvBool("ENABLE_ENCRYPTION", false)
}

// mustConnARPC creates an aRPC client with optional reliable delivery, congestion control, flow control, and encryption
// Features are controlled by ENABLE_RELIABLE, ENABLE_CC, ENABLE_FC, and ENABLE_ENCRYPTION environment variables
// (reliable/CC/FC default to true, encryption defaults to false)
func mustConnARPC(client **rpc.Client, addr string) {
	enableReliable := isReliableEnabled()
	enableCC := isCCEnabled()
	enableFC := isFCEnabled()
	enableEncryption := isEncryptionEnabled()

	log.Printf("Attempting to connect to aRPC server at: %s (reliable=%v, CC=%v, FC=%v, encryption=%v)",
		addr, enableReliable, enableCC, enableFC, enableEncryption)

	serializer := &serializer.SymphonySerializer{}

	clientLogger, logErr := messagelogger.NewClientMessageLogger()
	if logErr != nil {
		log.Printf("Failed to create client message logger: %v", logErr)
	}

	clientElements := []element.RPCElement{tracing.NewClientTracingElement(), clientLogger}

	var err error
	// Use 0.0.0.0:0 to explicitly bind to IPv4 instead of :0 which defaults to IPv6
	c, err := rpc.NewClientWithLocalAddr(serializer, addr, "0.0.0.0:0", clientElements)
	if err != nil {
		panic(errors.Wrapf(err, "arpc: failed to connect %s", addr))
	}

	// Get UDP transport from the client
	udpTransport := c.Transport()

	// Enable encryption if requested (independent of other features)
	if enableEncryption {
		udpTransport.EnableEncryption()
	}

	// If no other features are enabled, return basic client
	if !enableReliable && !enableCC && !enableFC {
		*client = c
		return
	}

	// Variables to hold handlers and packet types
	var reliableHandler *reliable.ReliableClientHandler
	var ccHandler *congestion.CCClientHandler
	var fcHandler *flowcontrol.FCClientHandler
	var ackPacketType packet.PacketType
	var ccFeedbackPacketType packet.PacketType
	var fcFeedbackPacketType packet.PacketType

	// Register packet types and create handlers based on enabled features
	if enableReliable {
		ackPacketType, err = udpTransport.RegisterPacketType(
			reliable.AckPacketName,
			&reliable.ACKPacketCodec{})
		if err != nil {
			panic(errors.Wrap(err, "failed to register ACK packet type"))
		}

		reliableHandler = reliable.NewReliableClientHandler(
			udpTransport,
			udpTransport.GetTimerManager(),
		)
	}

	if enableCC {
		ccFeedbackPacketType, err = udpTransport.RegisterPacketType(
			congestion.CCFeedbackPacketName,
			&congestion.CCFeedbackCodec{})
		if err != nil {
			panic(errors.Wrap(err, "failed to register CCFeedback packet type"))
		}

		ccHandler = congestion.NewCCClientHandler(
			udpTransport,
			udpTransport.GetTimerManager(),
		)
	}

	if enableFC {
		fcFeedbackPacketType, err = udpTransport.RegisterPacketType(
			flowcontrol.FCFeedbackPacketName,
			&flowcontrol.FCFeedbackCodec{})
		if err != nil {
			panic(errors.Wrap(err, "failed to register FCFeedback packet type"))
		}

		fcHandler = flowcontrol.NewFCClientHandler(
			udpTransport,
			udpTransport.GetTimerManager(),
		)
	}

	// Get existing handler chains for REQUEST packets (OnSend)
	requestChain, exists := udpTransport.GetHandlerRegistry().GetHandlerChain(
		packet.PacketTypeRequest.TypeID,
		transport.RoleClient,
	)
	if !exists {
		panic("failed to get REQUEST handler chain")
	}
	// Add handlers in order: CC, FC, then reliable
	if enableCC && ccHandler != nil {
		requestChain.AddHandler(ccHandler)
	}
	if enableFC && fcHandler != nil {
		requestChain.AddHandler(fcHandler)
	}
	if enableReliable && reliableHandler != nil {
		requestChain.AddHandler(reliableHandler)
	}

	// Get existing handler chains for RESPONSE packets (OnReceive)
	responseChain, exists := udpTransport.GetHandlerRegistry().GetHandlerChain(
		packet.PacketTypeResponse.TypeID,
		transport.RoleClient,
	)
	if !exists {
		panic("failed to get RESPONSE handler chain")
	}
	// Add handlers in order: reliable, CC, then FC
	if enableReliable && reliableHandler != nil {
		responseChain.AddHandler(reliableHandler)
	}
	if enableCC && ccHandler != nil {
		responseChain.AddHandler(ccHandler)
	}
	if enableFC && fcHandler != nil {
		responseChain.AddHandler(fcHandler)
	}

	// Register dedicated handler chains for feedback packets
	if enableReliable && reliableHandler != nil {
		ackChain := transport.NewHandlerChain("ClientACKHandlerChain", reliableHandler)
		udpTransport.RegisterHandlerChain(ackPacketType.TypeID, ackChain, transport.RoleClient)
	}

	if enableCC && ccHandler != nil {
		ccFeedbackChain := transport.NewHandlerChain("ClientCCFeedbackHandlerChain", ccHandler)
		udpTransport.RegisterHandlerChain(ccFeedbackPacketType.TypeID, ccFeedbackChain, transport.RoleClient)
	}

	if enableFC && fcHandler != nil {
		fcFeedbackChain := transport.NewHandlerChain("ClientFCFeedbackHandlerChain", fcHandler)
		udpTransport.RegisterHandlerChain(fcFeedbackPacketType.TypeID, fcFeedbackChain, transport.RoleClient)
	}

	*client = c
}

// setupServer configures an aRPC server with optional reliable delivery, congestion control, flow control, and encryption
// Features are controlled by ENABLE_RELIABLE, ENABLE_CC, ENABLE_FC, and ENABLE_ENCRYPTION environment variables
// (reliable/CC/FC default to true, encryption defaults to false)
// This function must be called after rpc.NewServer() but before server.Start()
// Returns a cleanup function that should be deferred
func setupServer(server *rpc.Server) func() {
	enableReliable := isReliableEnabled()
	enableCC := isCCEnabled()
	enableFC := isFCEnabled()
	enableEncryption := isEncryptionEnabled()

	log.Printf("Server configured with features: reliable=%v, CC=%v, FC=%v, encryption=%v",
		enableReliable, enableCC, enableFC, enableEncryption)

	// Get the UDP transport from the server
	udpTransport := server.GetTransport()

	// Enable encryption if requested (independent of other features)
	if enableEncryption {
		udpTransport.EnableEncryption()
	}

	// If no other features are enabled, return no-op cleanup function
	if !enableReliable && !enableCC && !enableFC {
		return func() {}
	}

	// Variables to hold handlers and packet types
	var reliableHandler *reliable.ReliableServerHandler
	var ccHandler *congestion.CCServerHandler
	var fcHandler *flowcontrol.FCServerHandler
	var ackPacketType packet.PacketType
	var ccFeedbackPacketType packet.PacketType
	var fcFeedbackPacketType packet.PacketType

	// Register packet types and create handlers based on enabled features
	if enableReliable {
		var err error
		ackPacketType, err = udpTransport.RegisterPacketType(reliable.AckPacketName, &reliable.ACKPacketCodec{})
		if err != nil {
			log.Fatalf("Failed to register ACK packet type: %v", err)
		}

		reliableHandler = reliable.NewReliableServerHandler(
			udpTransport,
			udpTransport.GetTimerManager(),
		)
	}

	if enableCC {
		var err error
		ccFeedbackPacketType, err = udpTransport.RegisterPacketType(congestion.CCFeedbackPacketName, &congestion.CCFeedbackCodec{})
		if err != nil {
			log.Fatalf("Failed to register CCFeedback packet type: %v", err)
		}

		ccHandler = congestion.NewCCServerHandler(
			udpTransport,
			udpTransport.GetTimerManager(),
		)
	}

	if enableFC {
		var err error
		fcFeedbackPacketType, err = udpTransport.RegisterPacketType(flowcontrol.FCFeedbackPacketName, &flowcontrol.FCFeedbackCodec{})
		if err != nil {
			log.Fatalf("Failed to register FCFeedback packet type: %v", err)
		}

		fcHandler = flowcontrol.NewFCServerHandler(
			udpTransport,
			udpTransport.GetTimerManager(),
		)
	}

	// Get handler chains for REQUEST packets (OnReceive)
	requestChain, exists := udpTransport.GetHandlerRegistry().GetHandlerChain(
		packet.PacketTypeRequest.TypeID,
		transport.RoleServer,
	)
	if !exists {
		log.Fatal("Failed to get REQUEST handler chain")
	}
	// Add handlers in order: reliable, CC, then FC
	if enableReliable && reliableHandler != nil {
		requestChain.AddHandler(reliableHandler)
	}
	if enableCC && ccHandler != nil {
		requestChain.AddHandler(ccHandler)
	}
	if enableFC && fcHandler != nil {
		requestChain.AddHandler(fcHandler)
	}

	// Get handler chains for RESPONSE packets (OnSend)
	responseChain, exists := udpTransport.GetHandlerRegistry().GetHandlerChain(
		packet.PacketTypeResponse.TypeID,
		transport.RoleServer,
	)
	if !exists {
		log.Fatal("Failed to get RESPONSE handler chain")
	}
	// Add handlers in order: CC, FC, then reliable
	if enableCC && ccHandler != nil {
		responseChain.AddHandler(ccHandler)
	}
	if enableFC && fcHandler != nil {
		responseChain.AddHandler(fcHandler)
	}
	if enableReliable && reliableHandler != nil {
		responseChain.AddHandler(reliableHandler)
	}

	// Register dedicated handler chains for feedback packets
	if enableReliable && reliableHandler != nil {
		ackChain := transport.NewHandlerChain("ServerACKHandlerChain", reliableHandler)
		udpTransport.RegisterHandlerChain(ackPacketType.TypeID, ackChain, transport.RoleServer)
	}

	if enableCC && ccHandler != nil {
		ccFeedbackChain := transport.NewHandlerChain("ServerCCFeedbackHandlerChain", ccHandler)
		udpTransport.RegisterHandlerChain(ccFeedbackPacketType.TypeID, ccFeedbackChain, transport.RoleServer)
	}

	if enableFC && fcHandler != nil {
		fcFeedbackChain := transport.NewHandlerChain("ServerFCFeedbackHandlerChain", fcHandler)
		udpTransport.RegisterHandlerChain(fcFeedbackPacketType.TypeID, fcFeedbackChain, transport.RoleServer)
	}

	// Return cleanup function that cleans up only the enabled handlers
	return func() {
		if enableReliable && reliableHandler != nil {
			reliableHandler.Cleanup()
		}
		if enableCC && ccHandler != nil {
			ccHandler.Cleanup()
		}
		if enableFC && fcHandler != nil {
			fcHandler.Cleanup()
		}
	}
}
