package services

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/appnet-org/arpc/pkg/logging"
	"github.com/appnet-org/arpc/pkg/rpc"
	pb "github.com/appnetorg/online-boutique-arpc/proto"
	"github.com/appnetorg/online-boutique-arpc/services/validator"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/pkg/errors"
)

const (
	defaultCurrency = "CNY"

	cookiePrefix   = "shop_"
	cookieCurrency = cookiePrefix + "currency"
)

type ctxKeySessionID struct{}
type ctxKeyRequestID struct{}

type platformDetails struct {
	css      string
	provider string
}

var (
	frontendMessage  = strings.TrimSpace(os.Getenv("FRONTEND_MESSAGE"))
	isCymbalBrand    = strings.ToLower(os.Getenv("CYMBAL_BRANDING")) == "true"
	assistantEnabled = strings.ToLower(os.Getenv("ENABLE_ASSISTANT")) == "true"
	templates        = template.Must(template.New("").
				Funcs(template.FuncMap{
			"renderMoney":        renderMoney,
			"renderCurrencyLogo": renderCurrencyLogo,
		}).ParseGlob("templates/*.html"))
	plat platformDetails

	whitelistedCurrencies = map[string]bool{
		"USD": true,
		"EUR": true,
		"CAD": true,
		"JPY": true,
		"GBP": true,
		"TRY": true,
	}
)

// frontendServer implements frontendServer service
type frontendServer struct {
	port int

	productCatalogSvcAddr string
	productCatalogSvcConn *rpc.Client

	currencySvcAddr string
	currencySvcConn *rpc.Client

	cartSvcAddr string
	cartSvcConn *rpc.Client

	recommendationSvcAddr string
	recommendationSvcConn *rpc.Client

	checkoutSvcAddr string
	checkoutSvcConn *rpc.Client

	shippingSvcAddr string
	shippingSvcConn *rpc.Client

	adSvcAddr string
	adSvcConn *rpc.Client

	shoppingAssistantSvcAddr string
}

func NewFrontendServer(port int) *frontendServer {
	return &frontendServer{
		port: port,
	}
}

// Run the server
func (fe *frontendServer) Run() error {
	err := logging.Init(getLoggingConfig())
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logging: %v", err))
	}

	mustMapEnv(&fe.productCatalogSvcAddr, "PRODUCT_CATALOG_SERVICE_ADDR")
	mustMapEnv(&fe.currencySvcAddr, "CURRENCY_SERVICE_ADDR")
	mustMapEnv(&fe.cartSvcAddr, "CART_SERVICE_ADDR")
	mustMapEnv(&fe.recommendationSvcAddr, "RECOMMENDATION_SERVICE_ADDR")
	mustMapEnv(&fe.checkoutSvcAddr, "CHECKOUT_SERVICE_ADDR")
	mustMapEnv(&fe.shippingSvcAddr, "SHIPPING_SERVICE_ADDR")
	mustMapEnv(&fe.adSvcAddr, "AD_SERVICE_ADDR")
	mustMapEnv(&fe.shoppingAssistantSvcAddr, "SHOPPING_ASSISTANT_SERVICE_ADDR")

	mustConnARPC(&fe.currencySvcConn, fe.currencySvcAddr)
	mustConnARPC(&fe.productCatalogSvcConn, fe.productCatalogSvcAddr)
	mustConnARPC(&fe.cartSvcConn, fe.cartSvcAddr)
	mustConnARPC(&fe.recommendationSvcConn, fe.recommendationSvcAddr)
	mustConnARPC(&fe.shippingSvcConn, fe.shippingSvcAddr)
	mustConnARPC(&fe.checkoutSvcConn, fe.checkoutSvcAddr)
	mustConnARPC(&fe.adSvcConn, fe.adSvcAddr)

	http.HandleFunc("/", fe.tracingMiddleware(fe.homeHandler))
	http.HandleFunc("/cart/checkout", fe.tracingMiddleware(fe.placeOrderHandler))
	http.HandleFunc("/cart", fe.tracingMiddleware(fe.addToCartHandler))

	log.Printf("frontendServer server running at port: %d", fe.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", fe.port), nil)
}

func (fe *frontendServer) tracingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracer := opentracing.GlobalTracer()
		spanName := fmt.Sprintf("HTTP %s %s", r.Method, r.URL.Path)

		span := tracer.StartSpan(spanName)
		defer span.Finish()

		// Set HTTP tags
		ext.HTTPMethod.Set(span, r.Method)
		ext.HTTPUrl.Set(span, r.URL.String())
		ext.Component.Set(span, "frontend")

		// Explicitly set service name
		span.SetTag("service.name", "frontend")
		span.SetTag("span.kind", "server")

		log.Printf("Created span: %s for service: frontend", spanName)

		// Add span to request context
		ctx := opentracing.ContextWithSpan(r.Context(), span)
		r = r.WithContext(ctx)

		// Call the next handler
		next(w, r)
	}
}

// homeHandler handles requests to the home page with detailed timing instrumentation
func (fe *frontendServer) homeHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.FormValue("user_id")

	log.Printf("homeHandler: Received request. UserID: %s, Currency: %s", userId, currentCurrency(r))

	// 1. Retrieve currencies
	currencies, err := fe.getCurrencies(r.Context(), userId)

	if err != nil {
		log.Printf("homeHandler: Error retrieving currencies: %v", err)
		renderHTTPError(r, w, errors.Wrap(err, "could not retrieve currencies"), http.StatusInternalServerError)
		return
	}
	log.Printf("homeHandler: Retrieved %d currencies", len(currencies))

	// 2. Retrieve products
	products, err := fe.getProducts(r.Context(), userId)

	if err != nil {
		log.Printf("homeHandler: Error retrieving products: %v", err)
		renderHTTPError(r, w, errors.Wrap(err, "could not retrieve products"), http.StatusInternalServerError)
		return
	}
	log.Printf("homeHandler: Retrieved %d products", len(products))

	// 3. Retrieve cart
	cart, err := fe.getCart(r.Context(), userId)

	if err != nil {
		log.Printf("homeHandler: Error retrieving cart: %v", err)
		renderHTTPError(r, w, errors.Wrap(err, "could not retrieve cart"), http.StatusInternalServerError)
		return
	}
	log.Printf("homeHandler: Retrieved cart with %d items", cartSize(cart))

	// 4. Process products for display with currency conversion
	type productView struct {
		Item  *pb.Product
		Price *pb.Money
	}
	ps := make([]productView, len(products))

	currencyConversionCount := 0
	currencyConversionErrors := 0

	for i, p := range products {
		price, err := fe.convertCurrency(r.Context(), p.GetPriceUsd(), currentCurrency(r), userId)

		if err != nil {
			currencyConversionErrors++
			log.Printf("homeHandler: Error converting currency for product %s: %v", p.GetId(), err)
			renderHTTPError(r, w, errors.Wrapf(err, "failed to do currency conversion for product %s", p.GetId()), http.StatusInternalServerError)
			return
		}

		currencyConversionCount++
		ps[i] = productView{p, price}
	}

	log.Printf("homeHandler: Processed %d products with currency conversion", len(ps))

	// 5. Get advertisement
	ad := fe.chooseAd(r.Context(), []string{}, userId)

	if ad != nil {
		log.Printf("homeHandler: Retrieved ad: %s", ad.GetRedirectUrl())
	}

	// 6. Render template
	err = templates.ExecuteTemplate(w, "home", injectCommonTemplateData(r, map[string]interface{}{
		"show_currency": true,
		"currencies":    currencies,
		"products":      ps,
		"cart_size":     cartSize(cart),
		"banner_color":  os.Getenv("BANNER_COLOR"), // illustrates canary deployments
		"ad":            ad,
	}))

	if err != nil {
		log.Printf("homeHandler: Error rendering template: %v", err)
	} else {
		log.Println("homeHandler: Successfully rendered home page")
	}
}

// placeOrderHandler handles placing an order
func (fe *frontendServer) placeOrderHandler(w http.ResponseWriter, r *http.Request) {
	// log.Println("placeOrderHandler: placing order")

	var (
		email         = r.FormValue("email")
		userId        = r.FormValue("user_id")
		streetAddress = r.FormValue("street_address")
		zipCode, _    = strconv.ParseInt(r.FormValue("zip_code"), 10, 32)
		city          = r.FormValue("city")
		state         = r.FormValue("state")
		country       = r.FormValue("country")
		ccNumber      = r.FormValue("credit_card_number")
		ccMonth, _    = strconv.ParseInt(r.FormValue("credit_card_expiration_month"), 10, 32)
		ccYear, _     = strconv.ParseInt(r.FormValue("credit_card_expiration_year"), 10, 32)
		ccCVV, _      = strconv.ParseInt(r.FormValue("credit_card_cvv"), 10, 32)
	)

	log.Printf("placeOrderHandler: received input - user_id: %s, email: %s, address: %s, city: %s, state: %s, country: %s, zip code: %d",
		userId, email, streetAddress, city, state, country, zipCode)

	payload := validator.PlaceOrderPayload{
		Email:         email,
		StreetAddress: streetAddress,
		ZipCode:       zipCode,
		City:          city,
		State:         state,
		Country:       country,
		CcNumber:      ccNumber,
		CcMonth:       ccMonth,
		CcYear:        ccYear,
		CcCVV:         ccCVV,
	}
	if err := payload.Validate(); err != nil {
		log.Printf("placeOrderHandler: validation error: %v", err)
		renderHTTPError(r, w, validator.ValidationErrorResponse(err), http.StatusUnprocessableEntity)
		return
	}
	log.Println("placeOrderHandler: input validation successful")

	checkoutClient := pb.NewCheckoutServiceClient(fe.checkoutSvcConn)
	order, err := checkoutClient.
		PlaceOrder(r.Context(), &pb.PlaceOrderRequest{
			Email: payload.Email,
			CreditCard: &pb.CreditCardInfo{
				CreditCardNumber:          payload.CcNumber,
				CreditCardExpirationMonth: int32(payload.CcMonth),
				CreditCardExpirationYear:  int32(payload.CcYear),
				CreditCardCvv:             int32(payload.CcCVV)},
			UserId:       sessionID(r),
			UserCurrency: currentCurrency(r),
			Address: &pb.Address{
				StreetAddress: payload.StreetAddress,
				City:          payload.City,
				State:         payload.State,
				ZipCode:       int32(payload.ZipCode),
				Country:       payload.Country},
		})
	if err != nil {
		log.Printf("placeOrderHandler: error placing order: %v", err)
		renderHTTPError(r, w, errors.Wrap(err, "failed to complete the order"), http.StatusInternalServerError)
		return
	}
	log.Printf("placeOrderHandler: order placed successfully, Order ID: %s", order.GetOrder().GetOrderId())

	recommendations, _ := fe.getRecommendations(r.Context(), sessionID(r), nil)
	log.Println("placeOrderHandler: retrieved recommendations")

	if len(recommendations) == 0 {
		log.Println("placeOrderHandler: No recommendations available")
	} else {
		for i, rec := range recommendations {
			log.Printf("Recommendation #%d: ID=%s, Name=%s, Description=%s, Picture=%s, PriceUSD=%v, Categories=%v",
				i+1, rec.Id, rec.Name, rec.Description, rec.Picture, rec.PriceUsd, rec.Categories)
		}
	}

	totalPaid := *order.GetOrder().GetShippingCost()
	for _, v := range order.GetOrder().GetItems() {
		multPrice := MultiplySlow(v.GetCost(), uint32(v.GetItem().GetQuantity()))
		totalPaid = *Must(Sum(&totalPaid, multPrice))
	}
	log.Printf("placeOrderHandler: total paid calculated: %d.%02d %s", totalPaid.GetUnits(), totalPaid.GetNanos()/10000000, totalPaid.GetCurrencyCode())

	currencies, err := fe.getCurrencies(r.Context(), userId)
	if err != nil {
		log.Printf("placeOrderHandler: error retrieving currencies: %v", err)
		renderHTTPError(r, w, errors.Wrap(err, "could not retrieve currencies"), http.StatusInternalServerError)
		return
	}
	log.Println("placeOrderHandler: retrieved currencies successfully")

	err = templates.ExecuteTemplate(w, "order", injectCommonTemplateData(r, map[string]interface{}{
		"show_currency":   false,
		"currencies":      currencies,
		"order":           order.GetOrder(),
		"total_paid":      &totalPaid,
		"recommendations": recommendations,
	}))
	if err != nil {
		log.Printf("placeOrderHandler: error rendering template: %v", err)
		return
	}
	log.Println("placeOrderHandler: order page rendered successfully")
}

func (fe *frontendServer) addToCartHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("addToCartHandler: Start processing request")

	quantity, _ := strconv.ParseUint(r.FormValue("quantity"), 10, 32)
	productID := r.FormValue("product_id")
	log.Printf("addToCartHandler: Received product_id=%s, quantity=%d", productID, quantity)

	payload := validator.AddToCartPayload{
		Quantity:  quantity,
		ProductID: productID,
	}

	// Validate payload
	if err := payload.Validate(); err != nil {
		log.Printf("addToCartHandler: Validation error for product_id=%s, quantity=%d: %v", productID, quantity, err)
		renderHTTPError(r, w, validator.ValidationErrorResponse(err), http.StatusUnprocessableEntity)
		return
	}
	log.Printf("addToCartHandler: Payload validated for product_id=%s, quantity=%d", productID, quantity)

	// Retrieve product details
	log.Printf("addToCartHandler: Fetching product details for product_id=%s", productID)
	p, err := fe.getProduct(r.Context(), payload.ProductID)
	if err != nil {
		log.Printf("addToCartHandler: Error retrieving product for product_id=%s: %v", productID, err)
		renderHTTPError(r, w, errors.Wrap(err, "could not retrieve product"), http.StatusInternalServerError)
		return
	}
	log.Printf("addToCartHandler: Retrieved product details for product_id=%s", productID)

	// Add product to cart
	log.Printf("addToCartHandler: Adding product_id=%s, quantity=%d to cart", productID, payload.Quantity)
	if err := fe.insertCart(r.Context(), sessionID(r), p.GetId(), int32(payload.Quantity)); err != nil {
		log.Printf("addToCartHandler: Error adding product_id=%s to cart: %v", productID, err)
		renderHTTPError(r, w, errors.Wrap(err, "failed to add to cart"), http.StatusInternalServerError)
		return
	}
	log.Printf("addToCartHandler: Successfully added product_id=%s, quantity=%d to cart", productID, payload.Quantity)

	// Redirect to cart
	w.Header().Set("location", "/cart")
	w.WriteHeader(http.StatusFound)
	log.Println("addToCartHandler: Redirected to /cart")
}

func (fe *frontendServer) getCurrencies(ctx context.Context, userID string) ([]string, error) {
	currencyClient := pb.NewCurrencyServiceClient(fe.currencySvcConn)
	currs, err := currencyClient.
		GetSupportedCurrencies(ctx, &pb.EmptyUser{UserId: userID})

	if err != nil {
		log.Printf("getCurrencies RPC failed: %v", err)
		return nil, err
	}

	var out []string
	for _, c := range currs.CurrencyCodes {
		if _, ok := whitelistedCurrencies[c]; ok {
			out = append(out, c)
		}
	}

	log.Printf("getCurrencies RPC completed, returned %d currencies", len(out))
	return out, nil
}

func (fe *frontendServer) getProducts(ctx context.Context, userID string) ([]*pb.Product, error) {
	productCatalogClient := pb.NewProductCatalogServiceClient(fe.productCatalogSvcConn)
	resp, err := productCatalogClient.
		ListProducts(ctx, &pb.EmptyUser{UserId: userID})

	if err != nil {
		log.Printf("getProducts RPC failed: %v", err)
		return nil, err
	}

	products := resp.GetProducts()
	log.Printf("getProducts RPC completed, returned %d products", len(products))
	return products, err
}

func (fe *frontendServer) getProduct(ctx context.Context, id string) (*pb.Product, error) {
	productCatalogClient := pb.NewProductCatalogServiceClient(fe.productCatalogSvcConn)
	resp, err := productCatalogClient.
		GetProduct(ctx, &pb.GetProductRequest{Id: id})
	return resp, err
}

func (fe *frontendServer) getCart(ctx context.Context, userID string) ([]*pb.CartItem, error) {
	cartClient := pb.NewCartServiceClient(fe.cartSvcConn)
	resp, err := cartClient.GetCart(ctx, &pb.GetCartRequest{UserId: userID})

	if err != nil {
		log.Printf("getCart RPC failed: %v", err)
		return nil, err
	}

	items := resp.GetItems()
	log.Printf("getCart RPC completed, returned %d items", len(items))
	return items, err
}

func (fe *frontendServer) insertCart(ctx context.Context, userID, productID string, quantity int32) error {
	cartClient := pb.NewCartServiceClient(fe.cartSvcConn)
	_, err := cartClient.AddItem(ctx, &pb.AddItemRequest{
		UserId: userID,
		Item: &pb.CartItem{
			ProductId: productID,
			Quantity:  quantity},
	})
	return err
}

func (fe *frontendServer) convertCurrency(ctx context.Context, money *pb.Money, currency string, userID string) (*pb.Money, error) {
	if money.GetCurrencyCode() == currency {
		return money, nil
	}

	currencyClient := pb.NewCurrencyServiceClient(fe.currencySvcConn)
	result, err := currencyClient.
		Convert(ctx, &pb.CurrencyConversionRequest{
			From:   money,
			ToCode: currency,
			UserId: userID})

	if err != nil {
		log.Printf("convertCurrency RPC failed: %v", err)
		return nil, err
	}

	log.Printf("convertCurrency RPC completed: %s -> %s", money.GetCurrencyCode(), currency)
	return result, err
}

func (fe *frontendServer) getRecommendations(ctx context.Context, userID string, productIDs []string) ([]*pb.Product, error) {
	recommendationClient := pb.NewRecommendationServiceClient(fe.recommendationSvcConn)
	resp, err := recommendationClient.ListRecommendations(ctx,
		&pb.ListRecommendationsRequest{UserId: userID, ProductIds: productIDs})
	if err != nil {
		return nil, err
	}
	out := make([]*pb.Product, len(resp.GetProductIds()))
	for i, v := range resp.GetProductIds() {
		p, err := fe.getProduct(ctx, v)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get recommended product info (#%s)", v)
		}
		out[i] = p
	}
	if len(out) > 4 {
		out = out[:4] // take only first four to fit the UI
	}
	return out, err
}

func (fe *frontendServer) getAd(ctx context.Context, ctxKeys []string, userID string) ([]*pb.Ad, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*100)
	defer cancel()

	adClient := pb.NewAdServiceClient(fe.adSvcConn)
	resp, err := adClient.GetAds(ctx, &pb.AdRequest{
		ContextKeys: ctxKeys,
		UserId:      userID,
	})

	if err != nil {
		log.Printf("getAd RPC failed: %v", err)
		return nil, errors.Wrap(err, "failed to get ads")
	}

	ads := resp.GetAds()
	log.Printf("getAd RPC completed, returned %d ads", len(ads))
	return ads, nil
}

func currentCurrency(r *http.Request) string {
	c, _ := r.Cookie(cookieCurrency)
	if c != nil {
		return c.Value
	}
	return defaultCurrency
}

func sessionID(r *http.Request) string {
	v := r.Context().Value(ctxKeySessionID{})
	if v != nil {
		return v.(string)
	}
	return ""
}

// renderHTTPError renders an error page and logs the error
func renderHTTPError(r *http.Request, w http.ResponseWriter, err error, code int) {
	log.Printf("renderHTTPError: request error: %v", err)

	errMsg := fmt.Sprintf("%+v", err)
	w.WriteHeader(code)

	// Attempt to render the error page
	templateErr := templates.ExecuteTemplate(w, "error", injectCommonTemplateData(r, map[string]interface{}{
		"error":       errMsg,
		"status_code": code,
		"status":      http.StatusText(code),
	}))
	if templateErr != nil {
		log.Printf("renderHTTPError: error rendering template: %v", templateErr)
	}
}

func renderMoney(money *pb.Money) string {
	currencyLogo := renderCurrencyLogo(money.GetCurrencyCode())
	return fmt.Sprintf("%s%d.%02d", currencyLogo, money.GetUnits(), money.GetNanos()/10000000)
}

func renderCurrencyLogo(currencyCode string) string {
	logos := map[string]string{
		"USD": "$",
		"CAD": "$",
		"JPY": "¥",
		"EUR": "€",
		"TRY": "₺",
		"GBP": "£",
	}

	logo := "$" //default
	if val, ok := logos[currencyCode]; ok {
		logo = val
	}
	return logo
}

func injectCommonTemplateData(r *http.Request, payload map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{
		"session_id":        sessionID(r),
		"request_id":        r.Context().Value(ctxKeyRequestID{}),
		"user_currency":     currentCurrency(r),
		"platform_css":      plat.css,
		"platform_name":     plat.provider,
		"is_cymbal_brand":   isCymbalBrand,
		"assistant_enabled": assistantEnabled,
		"frontendMessage":   frontendMessage,
		"currentYear":       time.Now().Year(),
	}

	for k, v := range payload {
		data[k] = v
	}

	return data
}

// get total # of items in cart
func cartSize(c []*pb.CartItem) int {
	cartSize := 0
	for _, item := range c {
		cartSize += int(item.GetQuantity())
	}
	return cartSize
}

// chooseAd queries for advertisements available and randomly chooses one, if
// available. It ignores the error retrieving the ad since it is not critical.
func (fe *frontendServer) chooseAd(ctx context.Context, ctxKeys []string, userId string) *pb.Ad {
	ads, err := fe.getAd(ctx, ctxKeys, userId)
	if err != nil {
		log.Printf("chooseAd: failed to retrieve ads: %v", err)
		return nil
	}

	// Choose a random ad from the retrieved ads
	return ads[rand.Intn(len(ads))]
}
