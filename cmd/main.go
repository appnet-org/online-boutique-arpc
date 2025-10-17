package main

import (
	"flag"
	"log"
	"os"

	services "github.com/appnetorg/online-boutique-arpc/services"
	"github.com/appnetorg/online-boutique-arpc/services/tracing"
	"github.com/opentracing/opentracing-go"
)

type server interface {
	Run() error
}

func main() {
	var (
		// port            = flag.Int("port", 11000, "The service port")
		frontendport       = flag.Int("frontendport", 11000, "frontend service port")
		cartport           = flag.Int("cartaddr", 11001, "cart service port")
		productcatalogport = flag.Int("productcatalogport", 11002, "productcatalog service port")
		currencyport       = flag.Int("currencyport", 11003, "currency service port")
		paymentport        = flag.Int("paymentport", 11004, "payment service port")
		shippingport       = flag.Int("shippingport", 11005, "shipping service port")
		emailport          = flag.Int("emailport", 11006, "email service port")
		checkoutport       = flag.Int("checkoutport", 11007, "checkout service port")
		recommendationport = flag.Int("recommendationport", 11008, "recommendation service port")
		adport             = flag.Int("adport", 11009, "ad service port")
	)
	flag.Parse()

	var srv server
	var cmd = os.Args[1]
	println("cmd parsed: ", cmd)

	tracer, closer, err := tracing.Init(cmd)
	if err != nil {
		log.Fatalf("ERROR: cannot init Jaeger: %v\n", err)
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)
	log.Printf("Jaeger Tracer Initialised for %s", cmd)

	switch cmd {
	case "cart":
		srv = services.NewCartService(*cartport)
	case "productcatalog":
		srv = services.NewProductCatalogService(*productcatalogport)
	case "currency":
		srv = services.NewCurrencyService(*currencyport)
	case "payment":
		srv = services.NewPaymentService(*paymentport)
	case "shipping":
		srv = services.NewShippingService(*shippingport)
	case "email":
		srv = services.NewEmailService(*emailport)
	case "checkout":
		srv = services.NewCheckoutService(*checkoutport)
	case "recommendation":
		srv = services.NewRecommendationService(*recommendationport)
	case "ad":
		srv = services.NewAdService(*adport)
	case "frontend":
		srv = services.NewFrontendServer(*frontendport)
	default:
		log.Fatalf("unknown cmd: %s", cmd)
	}

	if err := srv.Run(); err != nil {
		log.Fatalf("run %s error: %v", cmd, err)
	}
}
