module github.com/dennishilgert/apollo

go 1.21.5

replace github.com/firecracker-microvm/firecracker-go-sdk => ./third_party/firecracker-go-sdk

replace github.com/firecracker-microvm/firecracker-go-sdk/client/models => ./third_party/firecracker-go-sdk/client/models

require (
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/pflag v1.0.5
	google.golang.org/grpc v1.61.0
	google.golang.org/protobuf v1.32.0
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231106174013-bbf56f31fb17 // indirect
)
