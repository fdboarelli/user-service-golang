package client

import (
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
	"testing"
	"user/service/api"
)

func TestGetStatus(t *testing.T) {
	client, ctx, cancel, conn := CreateConnWithMetadata(t)
	defer CloseConnection(t, cancel, conn)
	response, err := client.GetStatus(ctx, &empty.Empty{})
	if err != nil {
		t.Fatalf("GRPC call failed: %v", err)
	}
	assert.Equal(t, api.ServiceStatus_UP, response.Status)
}
