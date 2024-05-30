# Getting Started

The process for setting up Apollo includes several steps.
Once you have placed the service binaries in their specific directories, you can proceed with the following steps.

## Downloading the Firecracker binary

To get the Firecracker binary you can either run the script `scripts/get-firecracker.sh` or build it from [source](https://github.com/firecracker-microvm/firecracker).

## Downloading a prebuilt Linux kernel

To get a Linux kernel you can either run the script `scripts/get-kernel.sh` or build one by yourself based on the [instructions](https://github.com/firecracker-microvm/firecracker/blob/main/docs/rootfs-and-kernel-setup.md) from the authors of Firecracker.

## Setting up CNI

To setup CNI run the script `scripts/setup-cni.sh`. This script will take care of installing the base CNI plugins together with the plugin for Firecracker.

## Setting up the environment

Each component supports providing the variables with a .env file or in the direct process environment.
Besides component specific variables, there some that can be applied for each service. These include:

```bash
# Defines the log level for the component.
APOLLO_LOG_LEVEL=info
```

### Fleet Manager

The Fleet Manager's environment configuration contains the following varrables:

```bash
# (Optional) Port on which the api server is listening for requests.
APOLLO_API_PORT=50051

# (Optional) Port on which the agent's api server is listening for requests.
APOLLO_AGENT_API_PORT=50051

# (Optional) Path where the persistent data like kernels, function code, etc. is stored at.
APOLLO_DATA_PATH=/data

# Absolute path to the Firecracker binary.
APOLLO_FC_BINARY_PATH=/home/user/fleet-manager/firecracker

# Network address of the image registry. This can be configured in the Docker compose file.
APOLLO_IMAGE_REGISTRY_ADDRESS=localhost:5000

# Comma-separated list of Kafka bootstrap servers. At least one bootstrap server address is necessary. This can be configured in the Docker compose file.
APOLLO_MESSAGING_BOOTSTRAP_SERVERS=localhost:9092,localhost:9093

# (Optional) Count of the worker go routines that handle messages from subscribed Kafka topics.
APOLLO_MESSAGING_WORKER_COUNT=10

# Network address of the object storage server. This can be configured in the Docker compose file.
APOLLO_STORAGE_ENDPOINT=localhost:9000

# Access key of the object storage server.
APOLLO_STORAGE_ACCESS_KEY_ID=key

# Secret access key of the object storage server.
APOLLO_STORAGE_SECRET_ACCESS_KEY=secret

# (Optional) Check interval of the runner watchdog in seconds.
APOLLO_WATCHDOG_CHECK_INTERVAL=3

# (Optional) Count of the worker go routines that handle the runner health checking.
APOLLO_WATCHDOG_WORKER_COUNT=10

# Network address of the Apollo service registry.
APOLLO_SERVICE_REGISTRY_ADDRESS=localhost:50051

# (Optional) Interval in which heartbeats are sent to the Apollo service registry in seconds.
APOLLO_HEARTBEAT_INTERVAL=3
```

### Frontend

The Frontend's environment configuration contains the following varrables:

```bash
# (Optional) Port on which the api server is listening for requests.
APOLLO_API_PORT=50051

# Comma-separated list of Kafka bootstrap servers. At least one bootstrap server address is necessary. This can be configured in the Docker compose file.
APOLLO_MESSAGING_BOOTSTRAP_SERVERS=localhost:9092,localhost:9093

# (Optional) Count of the worker go routines that handle messages from subscribed Kafka topics.
APOLLO_MESSAGING_WORKER_COUNT=10

# Network address of the Apollo service registry.
APOLLO_SERVICE_REGISTRY_ADDRESS=localhost:50051

# (Optional) Interval in which heartbeats are sent to the Apollo service registry in seconds.
APOLLO_HEARTBEAT_INTERVAL=3

# Host of the relational database. This can be configured in the Docker compose file.
APOLLO_DATABASE_HOST=localhost

# (Optional) Port of the relational database. This can be configured in the Docker compose file.
APOLLO_DATABASE_PORT=5432

# Username of the relational database. This can be configured in the Docker compose file.
APOLLO_DATABASE_USERNAME=apollo

# Password of the relational database. This can be configured in the Docker compose file.
APOLLO_DATABASE_PASSWORD=password

# (Optional) Name of the database within the relational database. This can be configured in the Docker compose file.
APOLLO_DATABASE_DB=apollo
```

### API Gateway

The API Gateway's environment configuration contains the following varrables:

```bash
# (Optional) Port on which the api server is listening for requests.
APOLLO_API_PORT=80

# Network address of the Apollo service registry.
APOLLO_SERVICE_REGISTRY_ADDRESS=localhost:50051
```

### Log Service

The Log Service's environment configuration contains the following varrables:

```bash
# (Optional) Port on which the api server is listening for requests.
APOLLO_API_PORT=50051

# Comma-separated list of Kafka bootstrap servers. At least one bootstrap server address is necessary. This can be configured in the Docker compose file.
APOLLO_MESSAGING_BOOTSTRAP_SERVERS=localhost:9092,localhost:9093

# (Optional) Count of the worker go routines that handle messages from subscribed Kafka topics.
APOLLO_MESSAGING_WORKER_COUNT=10

# Network address of the Apollo service registry.
APOLLO_SERVICE_REGISTRY_ADDRESS=localhost:50051

# (Optional) Interval in which heartbeats are sent to the Apollo service registry in seconds.
APOLLO_HEARTBEAT_INTERVAL=3

# Host of the document database. This can be configured in the Docker compose file.
APOLLO_DATABASE_HOST=localhost

# (Optional) Port of the document database. This can be configured in the Docker compose file.
APOLLO_DATABASE_PORT=5432

# Username of the document database. This can be configured in the Docker compose file.
APOLLO_DATABASE_USERNAME=apollo

# Password of the document database. This can be configured in the Docker compose file.
APOLLO_DATABASE_PASSWORD=password

# (Optional) Name of the database within the document database. This can be configured in the Docker compose file.
APOLLO_DATABASE_DB=apollo

# (Optional) Name of the database to authenticate against with the provided credentials. This can be configured in the Docker compose file.
APOLLO_DATABASE_AUTH_DB=admin
```

### Package Service

The Package Service's environment configuration contains the following varrables:

```bash
# (Optional) Port on which the api server is listening for requests.
APOLLO_API_PORT=50051

# Absolute path to the working directory of this service.
APOLLO_WORKING_DIR=/tmp/apollo

# (Optional) Count of the worker go routines that handle package creation requests.
APOLLO_PACKAGE_WORKER_COUNT=10

# Network address of the image registry. This can be configured in the Docker compose file.
APOLLO_IMAGE_REGISTRY_ADDRESS=localhost:5000

# Comma-separated list of Kafka bootstrap servers. At least one bootstrap server address is necessary. This can be configured in the Docker compose file.
APOLLO_MESSAGING_BOOTSTRAP_SERVERS=localhost:9092,localhost:9093

# (Optional) Count of the worker go routines that handle messages from subscribed Kafka topics.
APOLLO_MESSAGING_WORKER_COUNT=10

# Network address of the Apollo service registry.
APOLLO_SERVICE_REGISTRY_ADDRESS=localhost:50051

# (Optional) Interval in which heartbeats are sent to the Apollo service registry in seconds.
APOLLO_HEARTBEAT_INTERVAL=3

# Network address of the object storage server. This can be configured in the Docker compose file.
APOLLO_STORAGE_ENDPOINT=localhost:9000

# Access key of the object storage server.
APOLLO_STORAGE_ACCESS_KEY_ID=key

# Secret access key of the object storage server.
APOLLO_STORAGE_SECRET_ACCESS_KEY=secret
```

### Service Registry

The Service Registry's environment configuration contains the following varrables:

```bash
# (Optional) Port on which the api server is listening for requests.
APOLLO_API_PORT=50051

# Comma-separated list of Kafka bootstrap servers. At least one bootstrap server address is necessary. This can be configured in the Docker compose file.
APOLLO_MESSAGING_BOOTSTRAP_SERVERS=localhost:9092,localhost:9093

# (Optional) Count of the worker go routines that handle messages from subscribed Kafka topics.
APOLLO_MESSAGING_WORKER_COUNT=10

# Network address of the key/value store.
APOLLO_CACHE_ADDRESS=localhost:6379

# (Optional) Username for accessing the key/value store.
APOLLO_CACHE_USERNAME=username

# (Optional) Password for accessing the key/value store.
APOLLO_CACHE_PASSWORD=password

# (Optional) ID of the database inside of the key/value store.
APOLLO_CACHE_DATABASE=0
```

### Worker Manager

The Worker Manager's environment configuration contains the following varrables:

```bash
# (Optional) Port on which the api server is listening for requests.
APOLLO_API_PORT=50051

# Comma-separated list of Kafka bootstrap servers. At least one bootstrap server address is necessary. This can be configured in the Docker compose file.
APOLLO_MESSAGING_BOOTSTRAP_SERVERS=localhost:9092,localhost:9093

# (Optional) Count of the worker go routines that handle messages from subscribed Kafka topics.
APOLLO_MESSAGING_WORKER_COUNT=10

# Network address of the key/value store.
APOLLO_CACHE_ADDRESS=localhost:6379

# (Optional) Username for accessing the key/value store.
APOLLO_CACHE_USERNAME=username

# (Optional) Password for accessing the key/value store.
APOLLO_CACHE_PASSWORD=password

# (Optional) ID of the database inside of the key/value store.
APOLLO_CACHE_DATABASE=0

# Network address of the Apollo service registry.
APOLLO_SERVICE_REGISTRY_ADDRESS=localhost:50051

# (Optional) Interval in which heartbeats are sent to the Apollo service registry in seconds.
APOLLO_HEARTBEAT_INTERVAL=3

# (Optional) Factor for on how many worker nodes a function should be initialized.
APOLLO_FUNCTION_INITIALIZATION_FACTOR=3
```

## Building the resources

Apollo depends on some default resources like runtimes and initial function code in order to work properly. These need to be built manually by running `make resources-build`. Please note that the Docker Image Registry address variable must be exported to the environment by running `export APOLLO_IMAGE_REGISTRY_ADDRESS=your-address` before running this command.

## Setting up the third party services

In order to use the third party services for the needs of Apollo some configuration must be applied after these services have been started as Docker containers.

### Enable MinIO bucket notifications

Make sure the MinIO and Apache Kafka containers are up and running before running the script `scripts/setup-minio-notifications.sh`. This will create all necessary buckets and setup up the bucket notifications for Apollo.