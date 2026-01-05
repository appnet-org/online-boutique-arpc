package messagelogger

import (
	"encoding/json"
	"fmt"
	"reflect"

	pb "github.com/appnetorg/online-boutique-arpc/proto"
	"google.golang.org/protobuf/proto"
)

// SerializationSizes holds the sizes of different serialization formats
type SerializationSizes struct {
	Protobuf    int               `json:"protobuf"`
	FlatBuffers int               `json:"flatbuffers"`
	CapnProto   int               `json:"capnproto"`
	Symphony    int               `json:"symphony"`
	Errors      map[string]string `json:"errors,omitempty"`
}

// ComputeSizes computes serialization sizes for all formats
func ComputeSizes(msg interface{}) SerializationSizes {
	sizes := SerializationSizes{
		Protobuf:    -1,
		FlatBuffers: -1,
		CapnProto:   -1,
		Symphony:    -1,
		Errors:      make(map[string]string),
	}

	// Try to compute protobuf size
	if protoMsg, ok := msg.(proto.Message); ok {
		data, err := proto.Marshal(protoMsg)
		if err != nil {
			sizes.Errors["protobuf"] = err.Error()
		} else {
			sizes.Protobuf = len(data)
		}
	} else {
		sizes.Errors["protobuf"] = "not a proto message"
	}

	// Compute FlatBuffers, Cap'n Proto, and Symphony sizes based on message type
	fbData, fbErr := convertToFlatBuffers(msg)
	if fbErr != nil {
		sizes.Errors["flatbuffers"] = fbErr.Error()
	} else {
		sizes.FlatBuffers = len(fbData)
	}

	capnpData, capnpErr := convertToCapnProto(msg)
	if capnpErr != nil {
		sizes.Errors["capnproto"] = capnpErr.Error()
	} else {
		sizes.CapnProto = len(capnpData)
	}

	symphonyData, symphonyErr := convertToSymphony(msg)
	if symphonyErr != nil {
		sizes.Errors["symphony"] = symphonyErr.Error()
	} else {
		sizes.Symphony = len(symphonyData)
	}

	return sizes
}

// convertToFlatBuffers converts a protobuf message to FlatBuffers
func convertToFlatBuffers(msg interface{}) ([]byte, error) {
	typeName := reflect.TypeOf(msg).String()

	switch v := msg.(type) {
	case *pb.Empty:
		return ProtoToFB_Empty(v)
	case *pb.EmptyUser:
		return ProtoToFB_EmptyUser(v)
	case *pb.Money:
		return ProtoToFB_Money(v)
	case *pb.Address:
		return ProtoToFB_Address(v)
	case *pb.CartItem:
		return ProtoToFB_CartItem(v)
	case *pb.CreditCardInfo:
		return ProtoToFB_CreditCardInfo(v)
	case *pb.Ad:
		return ProtoToFB_Ad(v)
	case *pb.AddItemRequest:
		return ProtoToFB_AddItemRequest(v)
	case *pb.GetCartRequest:
		return ProtoToFB_GetCartRequest(v)
	case *pb.EmptyCartRequest:
		return ProtoToFB_EmptyCartRequest(v)
	case *pb.Cart:
		return ProtoToFB_Cart(v)
	case *pb.ListRecommendationsRequest:
		return ProtoToFB_ListRecommendationsRequest(v)
	case *pb.ListRecommendationsResponse:
		return ProtoToFB_ListRecommendationsResponse(v)
	case *pb.Product:
		return ProtoToFB_Product(v)
	case *pb.GetProductRequest:
		return ProtoToFB_GetProductRequest(v)
	case *pb.SearchProductsRequest:
		return ProtoToFB_SearchProductsRequest(v)
	case *pb.ListProductsResponse:
		return ProtoToFB_ListProductsResponse(v)
	case *pb.SearchProductsResponse:
		return ProtoToFB_SearchProductsResponse(v)
	case *pb.GetQuoteRequest:
		return ProtoToFB_GetQuoteRequest(v)
	case *pb.GetQuoteResponse:
		return ProtoToFB_GetQuoteResponse(v)
	case *pb.ShipOrderRequest:
		return ProtoToFB_ShipOrderRequest(v)
	case *pb.ShipOrderResponse:
		return ProtoToFB_ShipOrderResponse(v)
	case *pb.GetSupportedCurrenciesResponse:
		return ProtoToFB_GetSupportedCurrenciesResponse(v)
	case *pb.CurrencyConversionRequest:
		return ProtoToFB_CurrencyConversionRequest(v)
	case *pb.ChargeRequest:
		return ProtoToFB_ChargeRequest(v)
	case *pb.ChargeResponse:
		return ProtoToFB_ChargeResponse(v)
	case *pb.OrderItem:
		return ProtoToFB_OrderItem(v)
	case *pb.OrderResult:
		return ProtoToFB_OrderResult(v)
	case *pb.SendOrderConfirmationRequest:
		return ProtoToFB_SendOrderConfirmationRequest(v)
	case *pb.PlaceOrderRequest:
		return ProtoToFB_PlaceOrderRequest(v)
	case *pb.PlaceOrderResponse:
		return ProtoToFB_PlaceOrderResponse(v)
	case *pb.AdRequest:
		return ProtoToFB_AdRequest(v)
	case *pb.AdResponse:
		return ProtoToFB_AdResponse(v)
	default:
		return nil, fmt.Errorf("unsupported message type for FlatBuffers conversion: %s", typeName)
	}
}

// convertToCapnProto converts a protobuf message to Cap'n Proto
func convertToCapnProto(msg interface{}) ([]byte, error) {
	typeName := reflect.TypeOf(msg).String()

	switch v := msg.(type) {
	case *pb.Empty:
		return ProtoToCapnp_Empty(v)
	case *pb.EmptyUser:
		return ProtoToCapnp_EmptyUser(v)
	case *pb.Money:
		return ProtoToCapnp_Money(v)
	case *pb.Address:
		return ProtoToCapnp_Address(v)
	case *pb.CartItem:
		return ProtoToCapnp_CartItem(v)
	case *pb.CreditCardInfo:
		return ProtoToCapnp_CreditCardInfo(v)
	case *pb.Ad:
		return ProtoToCapnp_Ad(v)
	case *pb.AddItemRequest:
		return ProtoToCapnp_AddItemRequest(v)
	case *pb.GetCartRequest:
		return ProtoToCapnp_GetCartRequest(v)
	case *pb.EmptyCartRequest:
		return ProtoToCapnp_EmptyCartRequest(v)
	case *pb.Cart:
		return ProtoToCapnp_Cart(v)
	case *pb.ListRecommendationsRequest:
		return ProtoToCapnp_ListRecommendationsRequest(v)
	case *pb.ListRecommendationsResponse:
		return ProtoToCapnp_ListRecommendationsResponse(v)
	case *pb.Product:
		return ProtoToCapnp_Product(v)
	case *pb.GetProductRequest:
		return ProtoToCapnp_GetProductRequest(v)
	case *pb.SearchProductsRequest:
		return ProtoToCapnp_SearchProductsRequest(v)
	case *pb.ListProductsResponse:
		return ProtoToCapnp_ListProductsResponse(v)
	case *pb.SearchProductsResponse:
		return ProtoToCapnp_SearchProductsResponse(v)
	case *pb.GetQuoteRequest:
		return ProtoToCapnp_GetQuoteRequest(v)
	case *pb.GetQuoteResponse:
		return ProtoToCapnp_GetQuoteResponse(v)
	case *pb.ShipOrderRequest:
		return ProtoToCapnp_ShipOrderRequest(v)
	case *pb.ShipOrderResponse:
		return ProtoToCapnp_ShipOrderResponse(v)
	case *pb.GetSupportedCurrenciesResponse:
		return ProtoToCapnp_GetSupportedCurrenciesResponse(v)
	case *pb.CurrencyConversionRequest:
		return ProtoToCapnp_CurrencyConversionRequest(v)
	case *pb.ChargeRequest:
		return ProtoToCapnp_ChargeRequest(v)
	case *pb.ChargeResponse:
		return ProtoToCapnp_ChargeResponse(v)
	case *pb.OrderItem:
		return ProtoToCapnp_OrderItem(v)
	case *pb.OrderResult:
		return ProtoToCapnp_OrderResult(v)
	case *pb.SendOrderConfirmationRequest:
		return ProtoToCapnp_SendOrderConfirmationRequest(v)
	case *pb.PlaceOrderRequest:
		return ProtoToCapnp_PlaceOrderRequest(v)
	case *pb.PlaceOrderResponse:
		return ProtoToCapnp_PlaceOrderResponse(v)
	case *pb.AdRequest:
		return ProtoToCapnp_AdRequest(v)
	case *pb.AdResponse:
		return ProtoToCapnp_AdResponse(v)
	default:
		return nil, fmt.Errorf("unsupported message type for Cap'n Proto conversion: %s", typeName)
	}
}

// convertToSymphony converts a protobuf message to Symphony
func convertToSymphony(msg interface{}) ([]byte, error) {
	typeName := reflect.TypeOf(msg).String()

	// Define interface for Symphony marshallable types
	type symphonyMarshaller interface {
		MarshalSymphony() ([]byte, error)
	}

	// Check if the message implements MarshalSymphony
	if sm, ok := msg.(symphonyMarshaller); ok {
		return sm.MarshalSymphony()
	}

	return nil, fmt.Errorf("unsupported message type for Symphony conversion: %s", typeName)
}

// GetMessageTypeName returns the type name of the message
func GetMessageTypeName(msg interface{}) string {
	if msg == nil {
		return "unknown"
	}

	t := reflect.TypeOf(msg)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.Name()
}

// CreateLogEntry creates a structured log entry with serialization sizes
type LogEntry struct {
	Timestamp   string             `json:"timestamp"`
	Direction   string             `json:"direction"` // "request" or "response"
	Method      string             `json:"method,omitempty"`
	MessageType string             `json:"message_type"`
	Sizes       SerializationSizes `json:"sizes"`
	Payload     interface{}        `json:"payload"`
}

// MarshalLogEntry marshals a log entry to JSON
func MarshalLogEntry(entry LogEntry) ([]byte, error) {
	return json.Marshal(entry)
}
