# Getting Started Apollo Development Setup

This is a very simple quick-start guide that tells you how to set up the local development environment for Apollo.

## After cloning this project

### Initialize git submodules

The Apollo repository contains git submodules. A complete documentation on how to handle git submodules can be found in [the git docs](https://git-scm.com/book/fa/v2/Git-Tools-Submodules).

It is necessary to initialize and update the submodules with the following commands:

```bash
git submodule init
git submodule update
```

To update the submodule from its remote repository, run this command in the project root directory:

```bash
git submodule update --remote <submodule-name>
```

### Set up proto-compilation

Apollo uses gRPC as interprocess-communication. To define the service interfaces, Protocol Buffers are used.

Install Protocol Buffer compiler protoc
```bash
scripts/install-protoc.sh
```

Install necessary Go protoc plugins
```bash
make init-proto
```

Add go path to the system's path environment variable
```bash
export PATH=/home/user/go/bin:$PATH
```

Generate stubs from the .proto files
```bash
make gen-proto
```

