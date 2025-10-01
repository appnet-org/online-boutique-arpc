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
)

// NewRecommendationService returns a new server for the RecommendationService
func NewRecommendationService(port int, tracingElement element.RPCElement) *RecommendationService {
	return &RecommendationService{
		port:           port,
		tracingElement: tracingElement,
	}
}

// RecommendationService implements the RecommendationService
type RecommendationService struct {
	port int

	productCatalogSvcAddr string
	productCatalogClient  pb.ProductCatalogServiceClient

	tracingElement element.RPCElement
}

// Run starts the server
func (s *RecommendationService) Run() error {
	mustMapEnv(&s.productCatalogSvcAddr, "PRODUCT_CATALOG_SERVICE_ADDR")

	// Create ARPC client
	serializer := &serializer.SymphonySerializer{}
	rpcElements := []element.RPCElement{s.tracingElement}
	productCatalogClient, err := rpc.NewClient(serializer, s.productCatalogSvcAddr, rpcElements)
	if err != nil {
		log.Fatalf("Failed to create product catalog aRPC client: %v", err)
	}
	s.productCatalogClient = pb.NewProductCatalogServiceClient(productCatalogClient)

	// Create ARPC server
	server, err := rpc.NewServer("0.0.0.0:"+strconv.Itoa(s.port), serializer, rpcElements)
	if err != nil {
		log.Fatalf("Failed to start aRPC server: %v", err)
	}

	pb.RegisterRecommendationServiceServer(server, s)
	log.Printf("RecommendationService running at port: %d", s.port)
	server.Start()
	return nil
}

// ListRecommendations provides a list of recommended product IDs based on user and product history
func (s *RecommendationService) ListRecommendations(ctx context.Context, req *pb.ListRecommendationsRequest) (*pb.ListRecommendationsResponse, context.Context, error) {
	log.Printf("ListRecommendations request received for user_id = %v, product_ids = %v", req.GetUserId(), req.GetProductIds())

	// Fetch a list of products from the product catalog.
	catalogProducts, err := s.productCatalogClient.ListProducts(ctx, &pb.EmptyUser{UserId: req.GetUserId()})
	if err != nil {
		log.Printf("Error fetching catalog products: %v", err)
		return nil, ctx, err
	}

	// Remove user-provided products from the catalog to avoid recommending them.
	userProductIDs := req.GetProductIds()
	userIDs := make(map[string]struct{}, len(userProductIDs))
	for _, id := range userProductIDs {
		userIDs[id] = struct{}{}
	}

	filtered := make([]string, 0, len(catalogProducts.Products))
	for _, product := range catalogProducts.Products {
		if _, ok := userIDs[product.Id]; !ok {
			filtered = append(filtered, product.Id)
		}
	}

	// Sample from filtered products and return them.
	rand.Shuffle(len(filtered), func(i, j int) { filtered[i], filtered[j] = filtered[j], filtered[i] })

	const maxResponses = 5
	recommended := filtered
	if len(filtered) > maxResponses {
		recommended = filtered[:maxResponses]
	}

	return &pb.ListRecommendationsResponse{
		ProductIds: recommended,
	}, ctx, nil
}
