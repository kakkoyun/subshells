#!/bin/sh
set -xe

# This script runs the given program periodically, restarting it if it fails.

echo "Press [CTRL+C] to stop.."
while true; do
    "$@" 2>&1 & sleep 10; kill $!
    sleep 1
done
