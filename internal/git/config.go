package git

type GitConfig struct {
	URL    string
	Branch string
	Name   string
	Email  string
}

type GitCredentials struct {
	Username string
	Password string
}
