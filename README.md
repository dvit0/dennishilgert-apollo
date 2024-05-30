<picture>
   <source media="(prefers-color-scheme: dark)" srcset="docs/images/apollo-logo-text-transparent-bg-white.png">
   <source media="(prefers-color-scheme: light)" srcset="docs/images/apollo-logo-text-transparent-bg-black.png">
   <img alt="Apollo Logo" width="200" src="docs/images/apollo-logo-text-transparent-bg-white.png">
</picture>

A FaaS platform with the security of virtual machines and the light weight of containers. 

Built on top of [Firecracker](https://github.com/firecracker-microvm/firecracker) Micro VMs.

## What is Apollo?

Apollo is a cutting-edge Function as a Service (FaaS) platform, specifically engineered to harness the power of Firecracker, a leading open-source virtualization technology. Apollo leverages Firecracker's unique ability to create and manage microVMs — lightweight virtual machines optimized for security and efficiency. These microVMs are designed to run serverless functions, providing users with a secure, isolated environment that combines the best features of traditional VMs and containers. With Apollo, developers can deploy serverless applications that benefit from rapid scaling, reduced overhead, and an operational model that emphasizes speed and flexibility, all while maintaining a strong security posture.

## State of development

As Apollo is still under active development, not all components are finished and ready to use. The primary focus currently lies on finishing the core components to accomplish the use of the system. After the core components have been developed to a stable state, the feature components will follow.

To get an overview of the architecture of Apollo, you can check out the architecture diagram [here](docs/design.md#system-architecture).

A detailled description of each system component can be found [here](docs/design.md#system-components).

To understand what happens internally, take a look at the process flows [here](docs/process-flows.md).

You can check the state of the components either by checking the code or the following list.

### Core components

The core components of Apollo include all services the are necessary for the system to operate.

- **Apollo Agent** - State: done

- **Apollo Fleet Manager** - State: done

- **Apollo Worker Manager** - State: done

- **Apollo Frontend** - State: done

- **Apollo API Gateway** - State: done

- **Apollo Service Registry** - State: done

- **Apollo Package Service** - State: done

- **Apollo CLI** - State: work in progress

### Feature components

- **Apollo Log Service** - State: done

## Getting Started

To get started with Apollo, download the latest release binaries (coming soon) or build them from source.

You can build the Apollo components on any system that has Go 1.21.5+, the `make` utility and `bash` installed, as follows:

```bash
# This builds the binaries for the system architecture specified in the make file.
make build

# This builds the binaries for all supported system architecutes.
make cross-compile
```

Depending on the system architecture, the built binaries are placed under `bin/${os}_${arch}`.

Soon there will be an option to pull the Docker images for the different components directly from Docker Hub or build them by your own.

For now, each binary must be placed at a location of your choice. It is recommended to place every service binary inside of a separate folder as the environment variables can be set up with a specific `.env` file per service. The required environment variables per service and the next steps can be found [here](docs/getting-started.md).

## Capabilities & Features

Apollo has a small set of capabilities and features, which will be expanded in the future.

**Note**: Apollo still is limited to its core functionality and does not contain an implementation for handling users / teams yet. Furthermore there is no authentication present.

- It is open source, so feel free to contribute ;)

- Fully isolated execution of workloads ensuring the security of virtual machines and the light weight of containers.

- Function invocation can be triggered with a HTTP trigger.

- Synchronous invocation of functions.

- Support for one to many worker nodes to handle a high execution load.

- Dynamic and intelligent load balacing between all available worker nodes.

- Reuse of runners after a execution is done. Runners are **not** shared between different functions.

- Dynamic scaling of the whole system as it is built on the microservice architecture.

- Detection and elimination of unhealthy runners, services and worker nodes.

More to be added soon.

## Credits

Apollo reuses some small pieces of code written by [the Dappr authors](https://github.com/dapr/dapr).

## License

Project Apollo is under the MIT license. See the LICENSE file for details.