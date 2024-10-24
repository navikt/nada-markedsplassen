# Light-weight architecture decision: long running operations

|               |                                                                 |
|---------------|-----------------------------------------------------------------|
| Status        | Accepted                                                        |
| Authors       | Paul B. Beskow, Petter Jacobsen, Vebjørn Rekkebo, Erik Vattekar |
| Date proposed | 23.10.24                                                        |
| Date accepted | 24.10.24                                                        |

## Log

- 23.10.24 - Status: proposed
- 24.10.24 - Submitted for review
- 24.10.24 - Walkthrough
- 24.10.24 - We have decided to go for alternative (3) Transactional Queue (Postgres)
- 24.10.24 - Status: accepted

## Context

We have a few long-running operations, such as creating Google Cloud Workstations and mapping BigQuery 
datasets to Metabase. These operations orchestrate a lot of resources, take several minutes to complete and might 
fail under various scenarios, including:

1. Reach service quotas
2. Permission issues
3. Timeouts and transient failures
4. Synchronization and race conditions

Current Situation:

- Operations are not run as background processes, which limits user interaction on the frontend during execution.
- With the Metabase sync, the user will never know if the process succeeds or fails.
- Resource cleanup after failures is not always possible or consistent.
- Some resources cannot be deleted, leading to potential resource exhaustion.
- Quota limits may be reached without proper monitoring or alerts.
- Permission issues can halt operations and require manual intervention.
- Timeouts and transient failures require retry mechanisms.
- Only the leader will run the synchronization tasks, potentially underutilizing available resources.


## Decision to be made

We need to introduce a mechanism that addresses most, if not all, of these challenges while improving system 
reliability, resource utilization, and maintainability. The architecture should:

1. Handle long-running operations efficiently
2. Improve error handling and recovery mechanisms
3. Optimize resource utilization for asynchronous tasks
4. Add as little system complexity as possible
5. Not overload the developers with complex programming paradigms

## Alternatives

1. Do nothing
2. Command chaining pipeline patterns, with durable execution.
3. Transactional queue (Postgres)
4. Event-driven architecture
5. Workflow orchestration engine
6. Google Cloud Tasks

### Do nothing
- Maintains the status quo, means that we need to continue implementing idempotent services, and let the user refresh 
on the frontend until success.

### Command chaining pipeline patterns, with durable execution
- Offers a good balance between flexibility and structure for implementing workflows.
- Allows for easy composition of commands, which can be useful for complex operations.
- Can be adapted to handle long-running operations and retries.
- Relatively easy to implement and can reuse existing code.
- We will need to implement our own durable execution, using Postgres for storage

### Transactional queue (Postgres)
- Can import an existing library directly into our go, e.g., River
- We will need to migrate the schemas for these solutions, and add them to our database
- Compared to working with external job queues like Google PubSub or Redis, this means a simpler development model
- We reuse our Postgres database for durability
- No additional services need to be maintained

### Event-driven architecture
- Provides a flexible approach to handle complex workflows and state changes.
- Can address the need for expanded data catalog with push-based collection and event subscriptions, but not 
  something we are doing today
- Can be implemented with existing Go libraries and potentially reuse much of the current codebase.
- Likely too complex for what we need, too generic, very difficult to reason about
- We only have a few endpoints with long-running operations, so a bit overkill

### Workflow orchestration engine
- Best fits the requirements for durable execution model and handling long-running operations.
- Provides built-in support for workflows, which aligns with the system's main characteristics.
- Offers automatic retries, versioning, and visibility into workflow progress, addressing the issues of timeouts, failures, and long-running operations.
- Allows for easy implementation of state machines and complex workflows.
- Supports REST API integration.
- Can likely reuse much of the current codebase by implementing activities and workflows.
- Requires an additional service
- The programming model, while it has nice features, is also very complex to use

### Google Cloud Tasks
- Distributed task queue for asynchronous execution of work
- Supports automatic retries and dead-letter queues
- Integrates well with other Google Cloud services
- Scalable and serverless
- Allows for task scheduling and rate limiting
- Provides REST API for task management
- Can handle long-running operations through task chaining
- Supports task de-duplication for idempotency
- Offers visibility into task status and history 
- Requires setup of an external service
- Some uncertainty around how we perform the integration between Cloud Tasks <> Frontend <> Backend with regards to 
  authn/authz

## Recommendation

Make use of a transactional queue, like [River](https://riverqueue.com), which provides:

- Web interface (optional), for visibility into task status
- Automatic retries
- Unique jobs (args, status, period, etc.)
- Multiple isolated queues
- Configurable job timeouts (mark as completed, with errors, after some period of time)
- No additional dependencies
- Makes use of Golang generics, so we don't need to cast things all over the place

## Feedback

- We need some way of fetching running jobs based on the user ident, to get the best user experience
- Overall, no big concerns
- A long-term goal should be to move all scheduled jobs to river
- We also need to update the Metabase mapper to use this pattern, so the user does not need to refresh
- The library is relatively new, which means that it is not very mature, and we should expect breaking changes to the API
- We also need to handle schema migrations for River, on our own

## Addendum

- https://dagster.io/blog/skip-kafka-use-postgres-message-queue
