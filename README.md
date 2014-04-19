Pong
====

Pong is a program that generates http servers with predefined responses based on the configuration.
Pong is used to test proxy servers on different endpoint failure scenarios.

Installation:

```
make deps
make install
```

Usage

```
pong -c config.yaml
```

Examples:


Starts the server that responds ok 50% of requests, hangs for 1 second and responds 500 another 50% of the requests.

```yaml
servers:
  - addr: localhost:5000
    readtimeout: 10s
    writetimeout: 10s
    handlers:
      /:
        - rate: 50%
          code: 200
          body: ok
          delay: 0s

        - rate: 50%
          code: 500
          body: not ok
          delay: 1s
```
