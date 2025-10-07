package grpc_clients

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	pb "ride-sharing/shared/proto/payment"
)

type paymentServiceClient struct {
	Client pb.PaymentServiceClient
	conn   *grpc.ClientConn
}

func NewPaymentServiceClient() (*paymentServiceClient, error) {
	paymentServiceURL := os.Getenv("PAYMENT_SERVICE_URL")

	if paymentServiceURL == "" {
		paymentServiceURL = "payment-service:9094"
	}

	conn, err := grpc.NewClient(paymentServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewPaymentServiceClient(conn)

	return &paymentServiceClient{
		Client: client,
		conn:   conn,
	}, nil
}

func (c *paymentServiceClient) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return
		}
	}
}