package services

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/appnet-org/arpc/pkg/rpc"
	"github.com/appnet-org/arpc/pkg/rpc/element"
	"github.com/appnet-org/arpc/pkg/serializer"
	"github.com/google/uuid"

	pb "github.com/appnetorg/online-boutique-arpc/proto"
	"github.com/appnetorg/online-boutique-arpc/services/tracing"
)

type InvalidCreditCardErr struct{}

func (e InvalidCreditCardErr) Error() string {
	return "invalid credit card"
}

type UnacceptedCreditCardErr struct{}

func (e UnacceptedCreditCardErr) Error() string {
	return "credit card not accepted; only VISA or MasterCard are accepted"
}

type ExpiredCreditCardErr struct{}

func (e ExpiredCreditCardErr) Error() string {
	return "credit card expired"
}

func validateAndCharge(amount *pb.Money, card *pb.CreditCardInfo) (string, error) {
	// Perform some rudimentary validation.
	number := strings.ReplaceAll(card.CreditCardNumber, "-", "")
	var company string
	switch {
	case len(number) < 4:
		return "", InvalidCreditCardErr{}
	case number[0] == '4':
		company = "Visa"
	case number[0] == '5':
		company = "MasterCard"
	default:
		return "", UnacceptedCreditCardErr{}
	}

	if card.CreditCardCvv < 100 || card.CreditCardCvv > 9999 {
		return "", InvalidCreditCardErr{}
	}

	if time.Date(int(card.CreditCardExpirationYear), time.Month(card.CreditCardExpirationMonth), 0, 0, 0, 0, 0, time.Local).Before(time.Now()) {
		return "", ExpiredCreditCardErr{}
	}

	// Card is valid: process the transaction.
	log.Printf(
		"Transaction processed: company=%s, last_four=%s, currency=%s, amount=%d.%d",
		company,
		number[len(number)-4:],
		amount.CurrencyCode,
		amount.Units,
		amount.Nanos,
	)

	// Generate a transaction ID.
	return uuid.New().String(), nil
}

// NewPaymentService returns a new server for the PaymentService
func NewPaymentService(port int) *PaymentService {
	return &PaymentService{
		port: port,
	}
}

// PaymentService implements the PaymentService
type PaymentService struct {
	port int
}

// Run starts the server
func (s *PaymentService) Run() error {
	serializer := &serializer.SymphonySerializer{}
	rpcElements := []element.RPCElement{tracing.NewServerTracingElement()}
	server, err := rpc.NewServer("0.0.0.0:"+strconv.Itoa(s.port), serializer, rpcElements)
	if err != nil {
		log.Fatalf("Failed to start aRPC server: %v", err)
	}

	pb.RegisterPaymentServiceServer(server, s)
	log.Printf("PaymentService running at port: %d", s.port)
	server.Start()
	return nil
}

// Charge processes a payment charge request
func (s *PaymentService) Charge(ctx context.Context, req *pb.ChargeRequest) (*pb.ChargeResponse, context.Context, error) {
	log.Printf("Charge request received for amount: %v %v", req.GetAmount().GetCurrencyCode(), req.GetAmount().GetUnits())
	log.Printf("Credit Card Info: Number ending in ****%s, Expiry: %02d/%04d",
		req.GetCreditCard().GetCreditCardNumber()[len(req.GetCreditCard().GetCreditCardNumber())-4:],
		req.GetCreditCard().GetCreditCardExpirationMonth(),
		req.GetCreditCard().GetCreditCardExpirationYear())

	transactionID, err := validateAndCharge(req.GetAmount(), req.GetCreditCard())
	if err != nil {
		log.Printf("Transaction failed: %v", err)
		return nil, ctx, err
	}

	log.Printf("Transaction successful: %v", transactionID)

	return &pb.ChargeResponse{
		TransactionId: transactionID,
	}, ctx, nil
}
