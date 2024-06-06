package git

import "go.uber.org/zap"

type GitClient interface {
	LastCommit(options Options) (string, error)
	Clone(path string, options Options) (string, error)
}

type GitClientFactory struct {
}

func (f GitClientFactory) GetGitClient(logger *zap.SugaredLogger) GitClient {
	return NewMixedClient(logger)
}

type Options struct {
	URL       string
	Reference string
	Auth      *AuthOptions
}
