# go-amqp-tools
Go implementation of librabbitmq tools.

## Why?
Building librabbitmq and its tools for non-Linux platform(especially Windows) is not easy. So reimplementing them in Go as Go's amqp package doesn't have native library dependency and Go makes it super easy to build cross-platform CLIs.

# Tools
## amqp-consume
Consume messages from queue.
```
Usage:
  amqp-consume [OPTIONS]

Application Options:
  -u, --url=            the AMQP URL to connect to
  -s, --server=         the AMQP server to connect to
      --port=           the port to connect on (default: 5672)
      --vhost=          the vhost to use when connecting (default: /)
      --username=       the username to login with (default: guest)
      --password=       the password to login with (default: guest)
  -q, --queue=          the queue to consume from
  -e, --exchange=       bind the queue to this exchange
  -r, --routing-key=    the routing key to bind with
  -d, --declare         declare an exclusive queue (deprecated, use --exclusive instead)
  -x, --exclusive       declare the queue as exclusive
  -A, --no-ack          consume in no-ack mode
  -c, --count=          stop consuming after this many messages are consumed
  -p, --prefetch-count= receive only this many message at a time from the server

Help Options:
  -h, --help            Show this help message
```

## _Other tools are yet to be implemented_
