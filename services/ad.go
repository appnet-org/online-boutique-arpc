package services

import (
	"context"
	"log"
	"math/rand"
	"strconv"

	"github.com/appnet-org/arpc/pkg/rpc"
	"github.com/appnet-org/arpc/pkg/rpc/element"
	"github.com/appnet-org/arpc/pkg/serializer"

	pb "github.com/appnetorg/online-boutique-arpc/proto"
	"github.com/appnetorg/online-boutique-arpc/services/tracing"
)

const (
	maxAdsToServe = 2
)

// NewAdService returns a new server for the AdService
func NewAdService(port int) *AdService {
	return &AdService{
		port: port,
		ads:  createAdsMap(),
	}
}

// AdService implements the AdService
type AdService struct {
	port int
	ads  map[string]*pb.Ad
}

// Run starts the server
func (s *AdService) Run() error {
	rpcElements := []element.RPCElement{tracing.NewServerTracingElement()}
	serializer := &serializer.SymphonySerializer{}
	server, err := rpc.NewServer("0.0.0.0:"+strconv.Itoa(s.port), serializer, rpcElements)
	if err != nil {
		log.Fatalf("Failed to start aRPC server: %v", err)
	}

	pb.RegisterAdServiceServer(server, s)
	log.Printf("AdService running at port: %d", s.port)
	server.Start()
	return nil
}

// GetAds returns a list of ads based on the context keys
func (s *AdService) GetAds(ctx context.Context, req *pb.AdRequest) (*pb.AdResponse, context.Context, error) {
	log.Printf("GetAds request with context_keys = %v", req.GetContextKeys())

	var allAds []*pb.Ad
	keywords := req.GetContextKeys()

	if len(keywords) > 0 {
		for _, kw := range keywords {
			allAds = append(allAds, s.getAdsByCategory(kw)...)
		}
		if len(allAds) == 0 {
			// Serve random ads
			allAds = s.getRandomAds()
		}
	} else {
		allAds = s.getRandomAds()
	}

	return &pb.AdResponse{
		Ads: allAds,
	}, ctx, nil
}

func (s *AdService) getAdsByCategory(category string) []*pb.Ad {
	if adInstance, ok := s.ads[category]; ok {
		return []*pb.Ad{adInstance}
	}
	return nil
}

func (s *AdService) getRandomAds() []*pb.Ad {
	ads := make([]*pb.Ad, maxAdsToServe)
	vals := make([]*pb.Ad, 0, len(s.ads))
	for _, ad := range s.ads {
		vals = append(vals, ad)
	}
	for i := 0; i < maxAdsToServe; i++ {
		ads[i] = vals[rand.Intn(len(vals))]
	}
	return ads
}

func createAdsMap() map[string]*pb.Ad {
	return map[string]*pb.Ad{
		"hair": {
			RedirectUrl: "/product/2ZYFJ3GM2N",
			Text:        "Hairdryer for sale. 50% off.",
		},
		"clothing": {
			RedirectUrl: "/product/66VCHSJNUP",
			Text:        "Tank top for sale. 20% off.",
		},
		"accessories": {
			RedirectUrl: "/product/1YMWWN1N4O",
			Text:        "Watch for sale. Buy one, get second kit for free",
		},
		"footwear": {
			RedirectUrl: "/product/L9ECAV7KIM",
			Text:        "Loafers for sale. Buy one, get second one for free",
		},
		"decor": {
			RedirectUrl: "/product/0PUK6V6EV0",
			Text:        "Candle holder for sale. 30% off.",
		},
		"kitchen": {
			RedirectUrl: "/product/9SIQT8TOJO",
			Text:        "Bamboo glass jar for sale. 10% off.",
		},
	}
}
