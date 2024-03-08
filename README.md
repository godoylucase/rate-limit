# Notification Service

![gophers.jpg](.resources%2Fgophers.jpg)

## Table of Contents

| Section                                               | Description                                                                   |
|-------------------------------------------------------|-------------------------------------------------------------------------------|
| [Overview](#overview)                                 | A brief overview of the Notification Service.                                 |
| [Assumptions](#architecture)                          | What assumptions have been taken to build this library.                       |
| [Getting Started](#getting-started)                   | Instructions to set up and run the Notification Service.                      |
| [Rate Limiting Algorithms](#rate-limiting-algorithms) | An overview of the rate limiting algorithms used in the Notification Service. |

# Overview

A simple Notification Service implementation with rate limiting capabilities.

# Assumptions

- It uses redis as a data store for rate limiting. Meant to be used in a distributed environment.
- Configuration values for the service should be provided by the client via json file path location.
- Rate limited notifications are simply rejected, it is up to the client to handle the rejection whether to retry or
  not.
- Client must provide a `Gateway` interface implementation to send notifications. This is to allow the client to use
  their own notification service provider.

# Getting Started

Adding this library to your go module is as simple as running the following command:

```shell
go get github.com/godoylucase/rate-limit
```

There is a good example of how to use this library in
the [`example.go`](https://github.com/godoylucase/rate-limit/blob/develop/example.go) file at the root of the project.
In order to switch
between the sliding window (`sliding_window`) and fixed window algorithms (`fixed_window`), you can change
the `rate_limit.type` property in
the [`example_config.json`](https://github.com/godoylucase/rate-limit/blob/develop/example_config.json) file.

## Running the example (*)

This repository is equipped with a Makefile that has a target to run the example. To run the example, simply run the
following command:

```shell
make example
```

## Running tests (*)

To run the unit and integration tests simply run the following command:

```shell
make local-all
```

(*) Make sure you have installed `make`, `docker` and `docker-compose` in your machine for these to run.

# Rate Limiting Algorithms

Rate limiting is a crucial mechanism to control the rate of incoming requests to a system,
preventing abuse or overuse of resources. Two commonly used rate limiting algorithms are the
`Sliding Window` and `Fixed Window` algorithms. This library provides both implementations.

## Sliding Window

The sliding window algorithm is a rate-limiting technique that considers a moving time window to evaluate request rates.
It maintains a record of recent activities within a specified time window and allows or denies requests based on the
accumulated total of activities within that window.

### Operation

At each request, the algorithm checks the total of activities within the sliding window.
Expired activities are removed, and the current request is added to the total.
If the total exceeds the defined limit, the request is denied.

### Use Cases

Suitable for scenarios where a smooth and continuous rate limit is required.
Offers flexibility in defining the time window duration.

##### Pros:

- Smooth Limiting: Provides a smoother rate limit by considering a rolling time window.
- Flexibility: Allows customization of the time window duration based on specific use cases.
- Real-Time Sensitivity: Reacts more quickly to changes in traffic patterns due to its continuous evaluation.

#### Cons:

- Higher Memory Usage: May require more memory to store timestamped activities within the sliding window.
- Complexity: Implementing sliding window algorithms can be more complex due to continuous updates and removals.

## Fixed Window

The fixed window algorithm is a rate-limiting approach that divides time into fixed intervals and evaluates request
rates within those intervals. It resets the total at the end of each interval, allowing or denying requests based on the
accumulated total.

### Operation

At each request, the algorithm increments the total for the current interval.
When the interval ends, the total resets to zero.
If the total exceeds the defined limit, the request is denied.

### Use Cases

Suitable for scenarios where a strict, periodic rate limit is required.
Offers simplicity and predictability in rate limiting.

#### Pros

- Predictable Reset: Resets total at fixed intervals, providing a clear and predictable rate limiting pattern.
- Resource Efficiency: Typically requires less memory compared to sliding window algorithms.
- Ease of Implementation: Simpler to implement due to discrete and periodic evaluation intervals.

#### Cons

- Burstiness: May allow bursty traffic to exceed the limit at the beginning of each interval.
- Delayed Reaction: Reacts less quickly to sudden changes in traffic patterns as the evaluation occurs at fixed
  intervals.