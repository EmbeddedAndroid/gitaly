package operations

import (
	"os/exec"
	"testing"

	"google.golang.org/grpc/codes"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/internal/git/log"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
	"golang.org/x/net/context"

	pb "gitlab.com/gitlab-org/gitaly-proto/go"
)

func TestSuccessfulUserCreateBranchRequest(t *testing.T) {
	ctx, cancel := testhelper.Context()
	defer cancel()

	server := runOperationServiceServer(t)
	defer server.Stop()

	client, conn := newOperationClient(t)
	defer conn.Close()

	startPoint := "c7fbe50c7c7419d9701eebe64b1fdacc3df5b9dd"
	startPointCommit, err := log.GetCommit(ctx, testRepo, startPoint, "")
	require.NoError(t, err)
	user := &pb.User{
		Name:  []byte("Alejandro Rodríguez"),
		Email: []byte("alejandro@gitlab.com"),
	}

	testCases := []struct {
		desc           string
		branchName     []byte
		startPoint     string
		expectedBranch *pb.Branch
	}{
		{
			desc:       "valid branch",
			branchName: []byte("new-branch"),
			startPoint: startPoint,
			expectedBranch: &pb.Branch{
				Name:         []byte("new-branch"),
				TargetCommit: startPointCommit,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			branchName := testCase.branchName
			request := &pb.UserCreateBranchRequest{
				Repository: testRepo,
				BranchName: branchName,
				StartPoint: []byte(testCase.startPoint),
				User:       user,
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			response, err := client.UserCreateBranch(ctx, request)
			if testCase.expectedBranch != nil {
				defer exec.Command("git", "-C", testRepoPath, "branch", "-D", string(branchName)).Run()
			}

			require.NoError(t, err)
			require.Equal(t, testCase.expectedBranch, response.Branch)
		})
	}
}

func TestFailedUserCreateBranchRequest(t *testing.T) {
	server := runOperationServiceServer(t)
	defer server.Stop()

	client, conn := newOperationClient(t)
	defer conn.Close()

	user := &pb.User{
		Name:  []byte("Alejandro Rodríguez"),
		Email: []byte("alejandro@gitlab.com"),
	}
	testCases := []struct {
		desc       string
		branchName string
		startPoint string
		user       *pb.User
		code       codes.Code
	}{
		{
			desc:       "empty start_point",
			branchName: "shiny-new-branch",
			startPoint: "",
			user:       user,
			code:       codes.InvalidArgument,
		},
		{
			desc:       "empty user",
			branchName: "shiny-new-branch",
			startPoint: "master",
			user:       nil,
			code:       codes.InvalidArgument,
		},
		{
			desc:       "non-existing starting point",
			branchName: "new-branch",
			startPoint: "i-dont-exist",
			user:       user,
			code:       codes.FailedPrecondition,
		},

		{
			desc:       "branch exists",
			branchName: "master",
			startPoint: "master",
			user:       user,
			code:       codes.FailedPrecondition,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			request := &pb.UserCreateBranchRequest{
				Repository: testRepo,
				BranchName: []byte(testCase.branchName),
				StartPoint: []byte(testCase.startPoint),
				User:       testCase.user,
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			_, err := client.UserCreateBranch(ctx, request)
			testhelper.AssertGrpcError(t, err, testCase.code, "")
		})
	}
}
