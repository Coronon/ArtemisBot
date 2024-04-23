package git

type Repository interface {
	// Clean up the repository
	Close() error

	// Push an empty commit to the repository and return the commit hash
	PushEmptyCommit() (string, error)
}
