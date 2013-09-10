# Go wrapper for SysV Message Queues

sysv_mq is a Go wrapper for [SysV Message Queues](http://man7.org/linux/man-pages/man7/svipc.7.html). 

Documentation can be viewed at http://godoc.org/github.com/Shopify/sysv_mq.

sysv_mq is tested on Linux and OS X. To run the tests run `make test`. This
makes sure that any messages queues currently on your system are deleted before
running the tests.
