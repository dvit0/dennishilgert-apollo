package function

type Config struct {
	Runtime        string
	RuntimeVersion string
}

type Event struct {
	RequestType string
	Data        interface{}
}

type Context struct {
	Id          string
	Name        string
	Version     string
	MemoryLimit int
	Runtime     string
	Handler     string
}
