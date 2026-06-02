package manager

type Config struct {
	CachedClients ClientTokenIO
}


func DefaultConfig() Config {
	c := Config{}
	c.CachedClients = &DefaultClientTokenIO{
		ClientFolder: "tokens",
	}
	return c
}