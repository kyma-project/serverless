# Function Runtime Deprecation Schedule

This document outlines the planned deprecation and end-of-life (EOL) dates for supported Function runtimes in Kyma Serverless.

## Supported Runtimes and Deprecation Timeline

| Runtime | Planned Deprecation | Estimated EOL |
| --- | --- | --- |
| Node.js 22 | July 2026 | November 2026 |
| Node.js 24 | TBD | TBD |
| Python 3.12 | September 2026 | March 2027 |
| Python 3.14 | TBD | TBD |

## Deprecation History

### Node.js 20
- **Status**: Removed
- **Deprecation Version**: v1.10.0
- **Removal Version**: v1.14.0
- **Details**: Node.js 20 reached end-of-life. For more information, see [#2231](https://github.com/kyma-project/serverless/issues/2231) and [#2682](https://github.com/kyma-project/serverless/issues/2682). Migrate to Node.js 22 or newer.

> [!NOTE] 
> The deprecation and EOL dates listed in this document are **predictions based on current release cadence and Node.js/Python LTS schedules**. These dates are subject to change and may be adjusted based on the following:
> - Changes in the Kyma Serverless release schedule
> - Updates to upstream Node.js and Python LTS timelines
> - Community feedback and requirements
> - Security considerations
>
> Always check the [release notes](https://github.com/kyma-project/serverless/releases) for announcements regarding runtime deprecations and EOL timelines.

## Recommendations

- Plan upgrades to newer runtimes well in advance of deprecation dates
- Monitor release notes for any changes to this schedule
- For Functions using deprecated runtimes, migrate before the EOL date to avoid service disruption
