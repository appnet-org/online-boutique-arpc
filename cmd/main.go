package main

import (
	"flag"
	"log"
	"os"

	services "github.com/appnetorg/online-boutique-arpc/services"
	"github.com/appnetorg/online-boutique-arpc/services/tracing"
)

type server interface {
	Run() error
}

func main() {
	var (
		// port            = flag.Int("port", 8080, "The service port")
		frontendport       = flag.Int("frontendport", 8080, "frontend service port")
		cartport           = flag.Int("cartaddr", 8081, "cart service port")
		productcatalogport = flag.Int("productcatalogport", 8082, "productcatalog service port")
		currencyport       = flag.Int("currencyport", 8083, "currency service port")
		paymentport        = flag.Int("paymentport", 8084, "payment service port")
		shippingport       = flag.Int("shippingport", 8085, "shipping service port")
		emailport          = flag.Int("emailport", 8086, "email service port")
		checkoutport       = flag.Int("checkoutport", 8087, "checkout service port")
		recommendationport = flag.Int("recommendationport", 8088, "recommendation service port")
		adport             = flag.Int("adport", 8089, "ad service port")
	)
	flag.Parse()

	var srv server
	var cmd = os.Args[1]
	println("cmd parsed: ", cmd)

	tracingElement, err := tracing.NewTracingElement(cmd)
	if err != nil {
		log.Fatalf("ERROR: cannot init Jaeger: %v\n", err)
	}
	defer tracingElement.(*tracing.TracingElement).Close()
	log.Printf("Jaeger Tracer Initialised for %s", cmd)

	switch cmd {
	case "cart":
		srv = services.NewCartService(*cartport, tracingElement)
	case "productcatalog":
		srv = services.NewProductCatalogService(*productcatalogport, tracingElement)
	case "currency":
		srv = services.NewCurrencyService(*currencyport, tracingElement)
	case "payment":
		srv = services.NewPaymentService(*paymentport, tracingElement)
	case "shipping":
		srv = services.NewShippingService(*shippingport, tracingElement)
	case "email":
		srv = services.NewEmailService(*emailport, tracingElement)
	case "checkout":
		srv = services.NewCheckoutService(*checkoutport, tracingElement)
	case "recommendation":
		srv = services.NewRecommendationService(*recommendationport, tracingElement)
	case "ad":
		srv = services.NewAdService(*adport, tracingElement)
	case "frontend":
		srv = services.NewFrontendServer(*frontendport, tracingElement)
	default:
		log.Fatalf("unknown cmd: %s", cmd)
	}

	if err := srv.Run(); err != nil {
		log.Fatalf("run %s error: %v", cmd, err)
	}
}
