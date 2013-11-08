UNAME := $(shell uname -s)

test: clean
	go test -v

# This is to remove all queues before running the tests. This makes sure that if
# tests failed and leaked things to the queue, the queue is fresh before running
# the suite.
ifeq ($(UNAME), Linux)
clean:
	ipcs -q | grep -E -o "[0-9]{6,}" | xargs -L 1 -r ipcrm -q
endif

ifeq ($(UNAME), Darwin)
clean:
	ipcs -q | grep -E -o "[0-9]{6,}" | xargs -L 1 ipcrm -q
endif
