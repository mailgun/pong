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


Starts the server that response 200 and ok to all requests going to /hello

```yaml
servers:
  - addr: localhost:5000
    path: /hello
    readtimeout: 10s
    writetimeout: 10s
    handlers:
      /:
        100%:
          code: 200
          body: ok
          delay: 0s
          contenttype: text/plain
```
