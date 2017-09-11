package operations

import (
	pb "gitlab.com/gitlab-org/gitaly-proto/go"
	"gitlab.com/gitlab-org/gitaly/internal/rubyserver"
)

type server struct {
	*rubyserver.Server
}

// NewServer creates a new instance of a grpc OperationServiceServer
func NewServer(rs *rubyserver.Server) pb.OperationServiceServer {
	return &server{rs}
}
