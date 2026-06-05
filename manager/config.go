package manager

type Config struct {
	TokenStore TokenStore
}

func DefaultConfig() Config {
	return Config{
		TokenStore: &FileTokenStore{Dir: "tokens"},
	}
}
