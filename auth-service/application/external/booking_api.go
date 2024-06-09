package external

import (
	"context"
	booking "github.com/ZMS-DevOps/booking-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func NewBookingClient(address string) booking.BookingServiceClient {
	conn, err := getConnection(address)
	if err != nil {
		log.Fatalf("Failed to start gRPC connection to Catalogue service: %v", err)
	}
	return booking.NewBookingServiceClient(conn)
}

func getConnection(address string) (*grpc.ClientConn, error) {
	return grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func IfHostCanBeDeleted(bookingClient booking.BookingServiceClient, id string) (*booking.CheckDeleteHostResponse, error) {
	return bookingClient.CheckDeleteHost(
		context.TODO(),
		&booking.CheckDeleteHostRequest{
			HostId: id,
		})
}

func IfGuestCanBeDeleted(bookingClient booking.BookingServiceClient, id string) (*booking.CheckDeleteClientResponse, error) {
	return bookingClient.CheckDeleteClient(
		context.TODO(),
		&booking.CheckDeleteClientRequest{
			HostId: id,
		})
}
