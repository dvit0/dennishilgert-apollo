<picture>
   <source media="(prefers-color-scheme: dark)" srcset="docs/images/apollo-logo-text-transparent-bg-white.png">
   <source media="(prefers-color-scheme: light)" srcset="docs/images/apollo-logo-text-transparent-bg-black.png">
   <img alt="Apollo Logo" width="200" src="docs/images/apollo-logo-text-transparent-bg-white.png">
</picture>

A FaaS platform with the security of virtual machines and the light weight of containers. 

Based on [Firecracker](https://github.com/firecracker-microvm/firecracker) Micro-VM's.

## What is Apollo?

Apollo is a cutting-edge Function as a Service (FaaS) platform, specifically engineered to harness the power of Firecracker, a leading open-source virtualization technology. Apollo leverages Firecracker's unique ability to create and manage microVMs — lightweight virtual machines optimized for security and efficiency. These microVMs are designed to run serverless functions, providing users with a secure, isolated environment that combines the best features of traditional VMs and containers. With Apollo, developers can deploy serverless applications that benefit from rapid scaling, reduced overhead, and an operational model that emphasizes speed and flexibility, all while maintaining a strong security posture.

## Getting Started

To be done.

## Features & Capabilities

Apollo has a small set of features and capabilities, which will be expanded in the future.

**Note**: Apollo still is limited to its core functionality and does not contain an implementation for handling users / teams yet. Furthermore there is no authentication present.

The **Apollo CLI** can be used to:

- Build the Docker images for the function root filesystem
- Push the Docker images to the private Image Registry
- Manage the function lifecycle
   - Create a new function
   - Update a function
   - Delete a function
- Trigger a function execution

More to be added soon.

## Credits

Apollo reuses some small pieces of code written by [the Dappr authors](https://github.com/dapr/dapr).

## License

Project Apollo is under the MIT license. See the LICENSE file for details.