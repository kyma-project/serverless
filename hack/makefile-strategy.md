# Makefile architecture

The goal is to develop a way to extend Makefile targets in the most readable way, without keeping all targets in one file.

Pros of the architecture:

* targets are well organized
* single responsibility
* extensibility

## Dependencies description
* `Makefile` - The main makefile that allows for installing and running the Serverless. It's a high-level target to run the module without knowing its internals. It's the first contact point for the Serverless module users.
* `hack/Makefile` - High-level API that contains all targets that may be used by any CI/CD system. It has dependencies on the `hack/*.mk` makefiles.
* `hack/*.mk` - Contains common targets that may be used by other makefiles (they are included and shouldn't be run directly). Targets are grouped by functionality. They should contain helpers' targets.
* `components/operator/Makefile` - Contains all basic operations on Serverless Operator like builds, tests, etc., used during development. It's also used by `Makefile`.
* `components/serverless/Makefile` - Contains all basic operations on Serverless like builds, tests, etc., used during development.

## Good practices

Every makefile (`Makefile` and `*.mk`) must contain a few pieces, making the file more useful and human-readable:

* include `hack/help.mk` - this file provides the `help` target describing what is inside `Makefile` and what we can do with it.
* before `include` you must define the `PROJECT_ROOT` environment variable pointing to the project root directory.

Additionally, `Makefile` (but not `*.mk`) can also contain the following:

* Description - helps understand what the target does and shows it in the help. (`## description` after target name).
* Sections - allows for separations of targets based on their destination. (`##@`).

Example of target that includes all good practices:

```Makefile
PROJECT_ROOT=.
include ${PROJECT_ROOT}/hack/help.mk

##@ General

.PHONY: run
run: create-k3d install-serverless-main ## Create k3d cluster and install serverless from main
```
