package services

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/appnet-org/arpc/pkg/rpc"
	"github.com/appnet-org/arpc/pkg/rpc/element"
	"github.com/appnet-org/arpc/pkg/serializer"
	"github.com/redis/go-redis/v9"

	pb "github.com/appnetorg/online-boutique-arpc/proto"
	"github.com/appnetorg/online-boutique-arpc/services/tracing"
)

// NewCartService returns a new server for the CartService
func NewCartService(port int) *CartService {
	return &CartService{
		port: port,
	}
}

// CartService implements the CartService
type CartService struct {
	port int

	cartRedisAddr string
	rdb           *redis.Client // Redis client
}

// Run starts the server
func (s *CartService) Run() error {

	mustMapEnv(&s.cartRedisAddr, "CART_REDIS_ADDR")

	s.rdb = redis.NewClient(&redis.Options{
		Addr: s.cartRedisAddr,
	})

	serializer := &serializer.SymphonySerializer{}
	rpcElements := []element.RPCElement{tracing.NewServerTracingElement()}
	server, err := rpc.NewServer("0.0.0.0:"+strconv.Itoa(s.port), serializer, rpcElements)
	if err != nil {
		log.Fatalf("Failed to start aRPC server: %v", err)
	}

	pb.RegisterCartServiceServer(server, s)
	log.Printf("CartService running at port: %d", s.port)
	server.Start()
	return nil
}

// AddItem adds an item to the user's cart
func (s *CartService) AddItem(ctx context.Context, req *pb.AddItemRequest) (*pb.Empty, context.Context, error) {
	log.Printf("AddItem request for user_id = %v, product_id = %v, quantity = %v", req.GetUserId(), req.GetItem().GetProductId(), req.GetItem().GetQuantity())

	userID := req.GetUserId()
	item := req.GetItem()

	// Fetch the existing cart
	data, err := s.rdb.Get(ctx, userID).Result()
	var cart []*pb.CartItem
	if err == redis.Nil {
		cart = []*pb.CartItem{} // Empty cart
	} else if err != nil {
		log.Printf("Failed to fetch cart for user_id = %v: %v", userID, err)
		return nil, ctx, err
	} else {
		err = json.Unmarshal([]byte(data), &cart)
		if err != nil {
			log.Printf("Failed to unmarshal cart for user_id = %v: %v", userID, err)
			return nil, ctx, err
		}
	}

	// Add item to the cart
	cart = append(cart, item)

	// Save the updated cart
	cartData, err := json.Marshal(cart)
	if err != nil {
		log.Printf("Failed to marshal cart for user_id = %v: %v", userID, err)
		return nil, ctx, err
	}

	err = s.rdb.Set(ctx, userID, cartData, 0).Err()
	if err != nil {
		log.Printf("Failed to save cart for user_id = %v: %v", userID, err)
		return nil, ctx, err
	}

	return &pb.Empty{}, ctx, nil
}

// GetCart retrieves the cart for a user
func (s *CartService) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.Cart, context.Context, error) {
	log.Printf("GetCart request for user_id = %v", req.GetUserId())

	userID := req.GetUserId()
	data, err := s.rdb.Get(ctx, userID).Result()
	if err == redis.Nil {
		return &pb.Cart{
			UserId: userID,
			Items:  []*pb.CartItem{},
		}, ctx, nil
	} else if err != nil {
		log.Printf("Failed to fetch cart for user_id = %v: %v", userID, err)
		return nil, ctx, err
	}

	var cart []*pb.CartItem
	err = json.Unmarshal([]byte(data), &cart)
	if err != nil {
		log.Printf("Failed to unmarshal cart for user_id = %v: %v", userID, err)
		return nil, ctx, err
	}

	return &pb.Cart{
		UserId: userID,
		Items:  cart,
	}, ctx, nil
}

// EmptyCart clears the cart for a user
func (s *CartService) EmptyCart(ctx context.Context, req *pb.EmptyCartRequest) (*pb.Empty, context.Context, error) {
	log.Printf("EmptyCart request for user_id = %v", req.GetUserId())

	err := s.rdb.Del(ctx, req.GetUserId()).Err()
	if err != nil {
		log.Printf("Failed to delete cart for user_id = %v: %v", req.GetUserId(), err)
		return nil, ctx, err
	}

	return &pb.Empty{}, ctx, nil
}
