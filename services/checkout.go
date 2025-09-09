package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/appnet-org/arpc/pkg/rpc"
	"github.com/appnet-org/arpc/pkg/rpc/element"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/appnet-org/arpc/pkg/serializer"
	pb "github.com/appnetorg/online-boutique-arpc/proto"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	nanosMin = -999999999
	nanosMax = +999999999
	nanosMod = 1000000000
)

var (
	ErrInvalidValue        = errors.New("one of the specified money values is invalid")
	ErrMismatchingCurrency = errors.New("mismatching currency codes")
)

func init() {
	// Configure default log output
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.SetOutput(os.Stdout)
}

// NewCheckoutService returns a new server for the CheckoutService
func NewCheckoutService(port int) *CheckoutService {
	return &CheckoutService{
		port: port,
	}
}

// CheckoutService implements the CheckoutService
type CheckoutService struct {
	port int

	productCatalogSvcAddr string
	productCatalogClient  pb.ProductCatalogServiceClient

	cartSvcAddr string
	cartClient  pb.CartServiceClient

	currencySvcAddr string
	currencyClient  pb.CurrencyServiceClient

	shippingSvcAddr string
	shippingClient  pb.ShippingServiceClient

	emailSvcAddr string
	emailClient  pb.EmailServiceClient

	paymentSvcAddr string
	paymentClient  pb.PaymentServiceClient
}

// Run starts the server
func (cs *CheckoutService) Run() error {
	mustMapEnv(&cs.shippingSvcAddr, "SHIPPING_SERVICE_ADDR")
	mustMapEnv(&cs.productCatalogSvcAddr, "PRODUCT_CATALOG_SERVICE_ADDR")
	mustMapEnv(&cs.cartSvcAddr, "CART_SERVICE_ADDR")
	mustMapEnv(&cs.currencySvcAddr, "CURRENCY_SERVICE_ADDR")
	mustMapEnv(&cs.emailSvcAddr, "EMAIL_SERVICE_ADDR")
	mustMapEnv(&cs.paymentSvcAddr, "PAYMENT_SERVICE_ADDR")

	// Create ARPC clients
	serializer := &serializer.SymphonySerializer{}

	// Shipping client
	shippingClient, err := rpc.NewClient(serializer, cs.shippingSvcAddr, []element.RPCElement{})
	if err != nil {
		log.Fatalf("Failed to create shipping aRPC client: %v", err)
	}
	cs.shippingClient = pb.NewShippingServiceClient(shippingClient)

	// Product catalog client
	productCatalogClient, err := rpc.NewClient(serializer, cs.productCatalogSvcAddr, []element.RPCElement{})
	if err != nil {
		log.Fatalf("Failed to create product catalog aRPC client: %v", err)
	}
	cs.productCatalogClient = pb.NewProductCatalogServiceClient(productCatalogClient)

	// Cart client
	cartClient, err := rpc.NewClient(serializer, cs.cartSvcAddr, []element.RPCElement{})
	if err != nil {
		log.Fatalf("Failed to create cart aRPC client: %v", err)
	}
	cs.cartClient = pb.NewCartServiceClient(cartClient)

	// Currency client
	currencyClient, err := rpc.NewClient(serializer, cs.currencySvcAddr, []element.RPCElement{})
	if err != nil {
		log.Fatalf("Failed to create currency aRPC client: %v", err)
	}
	cs.currencyClient = pb.NewCurrencyServiceClient(currencyClient)

	// Email client
	emailClient, err := rpc.NewClient(serializer, cs.emailSvcAddr, []element.RPCElement{})
	if err != nil {
		log.Fatalf("Failed to create email aRPC client: %v", err)
	}
	cs.emailClient = pb.NewEmailServiceClient(emailClient)

	// Payment client
	paymentClient, err := rpc.NewClient(serializer, cs.paymentSvcAddr, []element.RPCElement{})
	if err != nil {
		log.Fatalf("Failed to create payment aRPC client: %v", err)
	}
	cs.paymentClient = pb.NewPaymentServiceClient(paymentClient)

	// Create ARPC server
	server, err := rpc.NewServer("0.0.0.0:"+strconv.Itoa(cs.port), serializer, []element.RPCElement{})
	if err != nil {
		log.Fatalf("Failed to start aRPC server: %v", err)
	}

	pb.RegisterCheckoutServiceServer(server, cs)
	log.Printf("CheckoutService running at port: %d", cs.port)
	server.Start()
	return nil
}

// PlaceOrder processes an order placement request
func (cs *CheckoutService) PlaceOrder(ctx context.Context, req *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, context.Context, error) {
	log.Printf("[PlaceOrder] user_id=%q user_currency=%q", req.UserId, req.UserCurrency)

	orderID, err := uuid.NewUUID()
	if err != nil {
		return nil, ctx, status.Errorf(codes.Internal, "failed to generate order uuid")
	}

	prep, err := cs.prepareOrderItemsAndShippingQuoteFromCart(ctx, req.UserId, req.UserCurrency, req.Address)
	if err != nil {
		return nil, ctx, status.Error(codes.Internal, err.Error())
	}

	total := pb.Money{CurrencyCode: req.UserCurrency,
		Units: 0,
		Nanos: 0}
	total = *Must(Sum(&total, prep.shippingCostLocalized))
	for _, it := range prep.orderItems {
		multPrice := MultiplySlow(it.Cost, uint32(it.GetItem().GetQuantity()))
		total = *Must(Sum(&total, multPrice))
	}

	txID, err := cs.chargeCard(ctx, &total, req.CreditCard)
	if err != nil {
		return nil, ctx, status.Errorf(codes.Internal, "failed to charge card: %+v", err)
	}
	log.Printf("payment went through (transaction_id: %s)", txID)

	shippingTrackingID, err := cs.shipOrder(ctx, req.Address, prep.cartItems)
	if err != nil {
		return nil, ctx, status.Errorf(codes.Unavailable, "shipping error: %+v", err)
	}

	_ = cs.emptyUserCart(ctx, req.UserId)

	orderResult := &pb.OrderResult{
		OrderId:            orderID.String(),
		ShippingTrackingId: shippingTrackingID,
		ShippingCost:       prep.shippingCostLocalized,
		ShippingAddress:    req.Address,
		Items:              prep.orderItems,
	}

	if err := cs.sendOrderConfirmation(ctx, req.Email, orderResult); err != nil {
		log.Printf("failed to send order confirmation to %q: %+v", req.Email, err)
	} else {
		log.Printf("order confirmation email sent to %q", req.Email)
	}
	resp := &pb.PlaceOrderResponse{Order: orderResult}
	return resp, ctx, nil
}

type orderPrep struct {
	orderItems            []*pb.OrderItem
	cartItems             []*pb.CartItem
	shippingCostLocalized *pb.Money
}

func (cs *CheckoutService) prepareOrderItemsAndShippingQuoteFromCart(ctx context.Context, userID, userCurrency string, address *pb.Address) (orderPrep, error) {
	log.Printf("prepareOrderItemsAndShippingQuoteFromCart: Start processing for userID=%s, userCurrency=%s", userID, userCurrency)

	var out orderPrep

	// Get user cart
	cartItems, err := cs.getUserCart(ctx, userID)
	if err != nil {
		log.Printf("prepareOrderItemsAndShippingQuoteFromCart: Error fetching cart for userID=%s: %v", userID, err)
		return out, fmt.Errorf("cart failure: %+v", err)
	}
	log.Printf("prepareOrderItemsAndShippingQuoteFromCart: Retrieved %d items from cart for userID=%s", len(cartItems), userID)

	// Prepare order items
	orderItems, err := cs.prepOrderItems(ctx, cartItems, userCurrency)
	if err != nil {
		log.Printf("prepareOrderItemsAndShippingQuoteFromCart: Error preparing order items for userID=%s: %v", userID, err)
		return out, fmt.Errorf("failed to prepare order: %+v", err)
	}
	log.Printf("prepareOrderItemsAndShippingQuoteFromCart: Prepared %d order items for userID=%s", len(orderItems), userID)

	// Quote shipping
	shippingUSD, err := cs.quoteShipping(ctx, address, cartItems)
	if err != nil {
		log.Printf("prepareOrderItemsAndShippingQuoteFromCart: Error quoting shipping for userID=%s: %v", userID, err)
		return out, fmt.Errorf("shipping quote failure: %+v", err)
	}
	log.Printf("prepareOrderItemsAndShippingQuoteFromCart: Received shipping quote in USD for userID=%s", userID)

	// Convert shipping cost
	shippingPrice, err := cs.convertCurrency(shippingUSD, userCurrency)
	if err != nil {
		log.Printf("prepareOrderItemsAndShippingQuoteFromCart: Error converting shipping cost to currency=%s for userID=%s: %v", userCurrency, userID, err)
		return out, fmt.Errorf("failed to convert shipping cost to currency: %+v", err)
	}
	log.Printf("prepareOrderItemsAndShippingQuoteFromCart: Converted shipping cost to currency=%s for userID=%s", userCurrency, userID)

	out.shippingCostLocalized = shippingPrice
	out.cartItems = cartItems
	out.orderItems = orderItems
	return out, nil
}

func (cs *CheckoutService) quoteShipping(ctx context.Context, address *pb.Address, items []*pb.CartItem) (*pb.Money, error) {
	shippingQuote, err := cs.shippingClient.GetQuote(ctx, &pb.GetQuoteRequest{
		Address: address,
		Items:   items})
	if err != nil {
		return nil, fmt.Errorf("failed to get shipping quote: %+v", err)
	}
	return shippingQuote.GetCostUsd(), nil
}

func (cs *CheckoutService) getUserCart(ctx context.Context, userID string) ([]*pb.CartItem, error) {
	cart, err := cs.cartClient.GetCart(ctx, &pb.GetCartRequest{UserId: userID})
	if err != nil {
		return nil, fmt.Errorf("failed to get user cart during checkout: %+v", err)
	}
	return cart.GetItems(), nil
}

func (cs *CheckoutService) emptyUserCart(ctx context.Context, userID string) error {
	if _, err := cs.cartClient.EmptyCart(ctx, &pb.EmptyCartRequest{UserId: userID}); err != nil {
		return fmt.Errorf("failed to empty user cart during checkout: %+v", err)
	}
	return nil
}

func (cs *CheckoutService) prepOrderItems(ctx context.Context, items []*pb.CartItem, userCurrency string) ([]*pb.OrderItem, error) {
	out := make([]*pb.OrderItem, len(items))
	cl := cs.productCatalogClient

	for i, item := range items {
		product, err := cl.GetProduct(ctx, &pb.GetProductRequest{Id: item.GetProductId()})
		if err != nil {
			return nil, fmt.Errorf("failed to get product #%q", item.GetProductId())
		}
		price, err := cs.convertCurrency(product.GetPriceUsd(), userCurrency)
		if err != nil {
			return nil, fmt.Errorf("failed to convert price of %q to %s", item.GetProductId(), userCurrency)
		}
		out[i] = &pb.OrderItem{
			Item: item,
			Cost: price}
	}
	return out, nil
}

func (cs *CheckoutService) convertCurrency(from *pb.Money, toCurrency string) (*pb.Money, error) {
	result, err := cs.currencyClient.Convert(context.TODO(), &pb.CurrencyConversionRequest{
		From:   from,
		ToCode: toCurrency})
	if err != nil {
		return nil, fmt.Errorf("failed to convert currency: %+v", err)
	}
	return result, err
}

func (cs *CheckoutService) chargeCard(ctx context.Context, amount *pb.Money, paymentInfo *pb.CreditCardInfo) (string, error) {
	paymentResp, err := cs.paymentClient.Charge(ctx, &pb.ChargeRequest{
		Amount:     amount,
		CreditCard: paymentInfo})
	if err != nil {
		return "", fmt.Errorf("could not charge the card: %+v", err)
	}
	return paymentResp.GetTransactionId(), nil
}

func (cs *CheckoutService) sendOrderConfirmation(ctx context.Context, email string, order *pb.OrderResult) error {
	_, err := cs.emailClient.SendOrderConfirmation(ctx, &pb.SendOrderConfirmationRequest{
		Email: email,
		Order: order})
	return err
}

func (cs *CheckoutService) shipOrder(ctx context.Context, address *pb.Address, items []*pb.CartItem) (string, error) {
	resp, err := cs.shippingClient.ShipOrder(ctx, &pb.ShipOrderRequest{
		Address: address,
		Items:   items})
	if err != nil {
		return "", fmt.Errorf("shipment failed: %+v", err)
	}
	return resp.GetTrackingId(), nil
}

// IsValid checks if specified value has a valid units/nanos signs and ranges.
func IsValid(m *pb.Money) bool {
	return signMatches(m) && validNanos(m.GetNanos())
}

func signMatches(m *pb.Money) bool {
	return m.GetNanos() == 0 || m.GetUnits() == 0 || (m.GetNanos() < 0) == (m.GetUnits() < 0)
}

func validNanos(nanos int32) bool { return nanosMin <= nanos && nanos <= nanosMax }

// IsZero returns true if the specified money value is equal to zero.
func IsZero(m *pb.Money) bool { return m.GetUnits() == 0 && m.GetNanos() == 0 }

// IsPositive returns true if the specified money value is valid and is
// positive.
func IsPositive(m *pb.Money) bool {
	return IsValid(m) && m.GetUnits() > 0 || (m.GetUnits() == 0 && m.GetNanos() > 0)
}

// IsNegative returns true if the specified money value is valid and is
// negative.
func IsNegative(m *pb.Money) bool {
	return IsValid(m) && m.GetUnits() < 0 || (m.GetUnits() == 0 && m.GetNanos() < 0)
}

// AreSameCurrency returns true if values l and r have a currency code and
// they are the same values.
func AreSameCurrency(l, r *pb.Money) bool {
	return l.GetCurrencyCode() == r.GetCurrencyCode() && l.GetCurrencyCode() != ""
}

// AreEquals returns true if values l and r are the equal, including the
// currency. This does not check validity of the provided values.
func AreEquals(l, r *pb.Money) bool {
	return l.GetCurrencyCode() == r.GetCurrencyCode() &&
		l.GetUnits() == r.GetUnits() && l.GetNanos() == r.GetNanos()
}

// Negate returns the same amount with the sign negated.
func Negate(m *pb.Money) pb.Money {
	return pb.Money{
		Units:        -m.GetUnits(),
		Nanos:        -m.GetNanos(),
		CurrencyCode: m.GetCurrencyCode()}
}

// Must panics if the given error is not nil. This can be used with other
// functions like: "m := Must(Sum(a,b))".
func Must(v *pb.Money, err error) *pb.Money {
	if err != nil {
		panic(err)
	}
	return v
}

// Sum adds two values. Returns an error if one of the values are invalid or
// currency codes are not matching (unless currency code is unspecified for
// both).
func Sum(l, r *pb.Money) (*pb.Money, error) {
	if !IsValid(l) || !IsValid(r) {
		return &pb.Money{}, ErrInvalidValue
	} else if l.GetCurrencyCode() != r.GetCurrencyCode() {
		return &pb.Money{}, ErrMismatchingCurrency
	}
	units := l.GetUnits() + r.GetUnits()
	nanos := l.GetNanos() + r.GetNanos()

	if (units == 0 && nanos == 0) || (units > 0 && nanos >= 0) || (units < 0 && nanos <= 0) {
		// same sign <units, nanos>
		units += int64(nanos / nanosMod)
		nanos = nanos % nanosMod
	} else {
		// different sign. nanos guaranteed to not to go over the limit
		if units > 0 {
			units--
			nanos += nanosMod
		} else {
			units++
			nanos -= nanosMod
		}
	}

	return &pb.Money{
		Units:        units,
		Nanos:        nanos,
		CurrencyCode: l.GetCurrencyCode()}, nil
}

// MultiplySlow is a slow multiplication operation done through adding the value
// to itself n-1 times.
func MultiplySlow(m *pb.Money, n uint32) *pb.Money {
	out := &pb.Money{
		Units:        m.GetUnits(),
		Nanos:        m.GetNanos(),
		CurrencyCode: m.GetCurrencyCode(),
	}
	for n > 1 {
		out = Must(Sum(out, m))
		n--
	}
	return out
}
