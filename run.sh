#!/bin/sh

set -x

# This makes sure we don't spend a lot of money on API calls when the
# process starts crash looping for whatever reason:
./usr/local/bin/cost-exporter "$@" &

while true; do sleep 365d; done
