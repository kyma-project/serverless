# Makefile architecture

The goal is to develop a way to extend Makefile targets in the most readable way, without keeping all targets in one file.

Pros of the architecture:

* targets are well organized
* single responsibility
* extensibility

## Dependencies description
* `Makefile` - The main makefile which allow install and run serverless on local k3d. It should be readable for every person that doesn't know this article.
* `hack/Makefile` - High-level API that contains all targets that may be used by any CI/CD system. It has dependencies on the `hack/*.mk` makefiles.
* `hack/*.mk` - Contains common targets that may be used by other makefiles (they are included and shouldn't be run directly). Targets are groupped by functionality. They should contain helpers' targets.
* `components/operator/Makefile` - Contains all basic operations on serverless operator like builds, tests, etc. used during development. It's also used by `Makefile`.
* `components/serverless/Makefile` - Contains all basic operations on serverless like builds, tests, etc. used during development. It's used by `Makefile`.

## Good practices

Every makefile (`Makefile` and `*.mk`) must contain a few pieces, making the file more useful and human-readable:

* include `hack/help.mk` - this file provide `help` target describing what is inside Makefile and what we can do with it.
* before any `include` you should define `PROJECT_ROOT` environment variable pointing on project root directory.

Additionaly `Makefile` (but not `*.mk`) should contain:

* Description - helps understand what the target does and shows it in the help. (`## description` after target name).
* Sections - allow for separations of targets based on their destination. (`##@`).

Example of target that includes all good practices:

```Makefile
PROJECT_ROOT=.
include ${PROJECT_ROOT}/hack/help.mk

##@ General

.PHONY: run
run: create-k3d install-serverless-main ## Create k3d cluster and install serverless from main
```