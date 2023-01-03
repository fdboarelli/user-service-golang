package client

import (
	"context"
	"google.golang.org/grpc"
	"testing"
	"time"
	"user/service/api"
)

func CreateConnWithMetadata(t *testing.T) (api.UserServiceClient, context.Context, context.CancelFunc, *grpc.ClientConn) {
	// set up a connection to the server.
	conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("could not connect: %v", err)
	}
	client := api.NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	return client, ctx, cancel, conn
}

func CloseConnection(t *testing.T, cancel context.CancelFunc, conn *grpc.ClientConn) {
	cancel()
	err := conn.Close()
	if err != nil {
		t.Fatalf("failed to close connection - %s", err.Error())
	}
}
