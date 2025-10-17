package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"strconv"

	"github.com/appnet-org/arpc/pkg/logging"
	"github.com/appnet-org/arpc/pkg/rpc"
	"github.com/appnet-org/arpc/pkg/rpc/element"
	"github.com/appnet-org/arpc/pkg/serializer"

	pb "github.com/appnetorg/online-boutique-arpc/proto"
	"github.com/appnetorg/online-boutique-arpc/services/tracing"
)

// Embed the HTML template for the email
var (
	tmpl = template.Must(template.New("email").
		Funcs(template.FuncMap{
			"div": func(x, y int32) int32 { return x / y },
		}).
		Parse("./templates/email.html"))
)

// NewEmailService returns a new server for the EmailService
func NewEmailService(port int) *EmailService {
	return &EmailService{
		port: port,
	}
}

// EmailService implements the EmailService
type EmailService struct {
	port int
}

// Run starts the server
func (s *EmailService) Run() error {
	err := logging.Init(getLoggingConfig())
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logging: %v", err))
	}

	rpcElements := []element.RPCElement{tracing.NewServerTracingElement()}
	serializer := &serializer.SymphonySerializer{}
	server, err := rpc.NewServer("0.0.0.0:"+strconv.Itoa(s.port), serializer, rpcElements)
	if err != nil {
		log.Fatalf("Failed to start aRPC server: %v", err)
	}

	pb.RegisterEmailServiceServer(server, s)
	log.Printf("EmailService running at port: %d", s.port)
	server.Start()
	return nil
}

// SendOrderConfirmation sends an order confirmation email
func (s *EmailService) SendOrderConfirmation(ctx context.Context, req *pb.SendOrderConfirmationRequest) (*pb.Empty, context.Context, error) {
	log.Printf("SendOrderConfirmation request received for email = %v", req.GetEmail())

	// Generate email content using the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, req.GetOrder()); err != nil {
		log.Printf("Error executing template: %v", err)
		return nil, ctx, err
	}
	confirmation := buf.String()

	// Simulate sending the email
	log.Printf("Order confirmation email content for %v:\n%s", req.GetEmail(), confirmation)

	// Replace this with actual email-sending logic if needed
	log.Printf("Order confirmation email sent to %v", req.GetEmail())

	return &pb.Empty{}, ctx, nil
}
