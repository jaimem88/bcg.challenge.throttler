# BCG Coding Challenge

Implementation of an API Service Throttler

## Spec

- Start by creating a service with an endpoint to retrieve a list of users. This endpoint can simply return an empty array in the response.
- Every request should include an access token, otherwise it should fail. Any non-empty string is a valid token, otherwise it should fail as well.
- For each token, restrict access to N requests per M milliseconds where N and M are configurable. Once the limit has been reached, subsequent requests should fail and provide the amount of milliseconds remaining until reset to allow for rescheduling.

## Assumptions

1. Token must be present in the `Authorization` header as a `Bearer` token type
2. Used an unsafe in-memory cache using a map. An external service, e.g. Redis, could be used to persist the cache
3. Using start/end times to check limits. I could have used a `time.Timer` listening for the `channel` to update the in memory cache.

## Running

Runs with basic config in [cmd/bcg.service.throttler/config.go](cmd/bcg.service.throttler/config.go).
Optionally, a JSON config file can be passed via the `-config` flag

```sh
go build ./cmd/bcg.service.throttler/ && ./bcg.service.throttler -config config.json
```

## Unit tests

**NOTE**
One of the testing scenarios sleeps for a few milliseconds while waiting for the limit to reset, so they whole suite takes a bit longer to run.

```sh
go test
```

Or in a Docker container

```sh
make test
```

## Build docker image

Creates a config file, builds the binary and builds a Docker image.

```sh
make build
```

## Running in a Docker container

Requires $PORT to expose it in the running machine

```sh
make run PORT=8080
```