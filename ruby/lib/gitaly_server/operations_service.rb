module GitalyServer
  class OperationsService < Gitaly::OperationService::Service
    include Utils

    def user_create_branch(request, _call)
      repo = Gitlab::Git::Repository.from_call(_call)
      target = request.start_point
      raise GRPC::InvalidArgument, 'empty start_point' if target.empty?
      user = request.user
      raise GRPC::InvalidArgument, 'empty user' unless user

      target_object = Gitlab::Git::Ref.dereference_object(repo.lookup(target))
      branch_name = request.branch_name
      committer = Gitlab::Git::Committer.new(user.name, user.email, user.gl_id)
      Gitlab::Git::OperationService.new(committer, repo)
        .add_branch(branch_name, target_object.oid)
      created_branch = repo.find_branch(branch_name)
      return Gitaly::UserCreateBranchResponse.new unless created_branch

      commit = gitaly_commit_from_rugged(target_object)
      branch = Gitaly::Branch.new(name: branch_name, target_commit: commit)
      Gitaly::UserCreateBranchResponse.new(branch: branch)
    rescue Rugged::ReferenceError, Gitlab::Git::CommitError => ex
      raise GRPC::FailedPrecondition, ex.message
    end
  end
end
