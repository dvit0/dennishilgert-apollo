# Getting Started Apollo Development Setup

This is a very simple quick-start guide that tells you how to set up the local development environment for Apollo.

## After cloning this project
This repository contains git submodules. A complete documentation on how to handle git submodules can be found in [the git docs](https://git-scm.com/book/fa/v2/Git-Tools-Submodules).

It is necessary to initialize and update the submodules with the following commands:

```bash
git submodule init
git submodule update
```

To update the submodule from its remote repository, run this command in the project root directory:

```bash
git submodule update --remote <submodule-name>
```

