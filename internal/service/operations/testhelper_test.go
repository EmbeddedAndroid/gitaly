package operations

import (
	"net"
	"os"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"

	pb "gitlab.com/gitlab-org/gitaly-proto/go"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/internal/rubyserver"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	serverSocketPath = testhelper.GetTemporaryGitalySocketFileName()
	testRepo         *pb.Repository
	testRepoPath     string
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

var rubyServer *rubyserver.Server

func testMain(m *testing.M) int {
	defer testhelper.MustHaveNoChildProcess()

	testRepo = testhelper.TestRepository()

	var err error
	testRepoPath, err = helper.GetRepoPath(testRepo)
	if err != nil {
		log.Fatal(err)
	}

	testhelper.ConfigureRuby()
	rubyServer, err = rubyserver.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer rubyServer.Stop()

	return m.Run()
}

func runOperationServiceServer(t *testing.T) *grpc.Server {
	os.Remove(serverSocketPath)
	grpcServer := testhelper.NewTestGrpcServer(t, nil, nil)

	listener, err := net.Listen("unix", serverSocketPath)
	if err != nil {
		t.Fatal(err)
	}

	pb.RegisterOperationServiceServer(grpcServer, &server{rubyServer})
	reflection.Register(grpcServer)

	go grpcServer.Serve(listener)

	return grpcServer
}

func newOperationClient(t *testing.T) (pb.OperationServiceClient, *grpc.ClientConn) {
	connOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDialer(func(addr string, _ time.Duration) (net.Conn, error) {
			return net.Dial("unix", addr)
		}),
	}
	conn, err := grpc.Dial(serverSocketPath, connOpts...)
	if err != nil {
		t.Fatal(err)
	}

	return pb.NewOperationServiceClient(conn), conn
}
