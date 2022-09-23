#!/usr/bin/env sh
set -e
# Install packr2 and generate packr boxes for Polaris.
# IDeally this script is called with $GOBIN already set to a temporary
# directory, where packr2 will be installed.
if [ "x${GOBIN}" != "x" ] ; then
  PATH=$GOBIN:$PATH
fi
go install github.com/gobuffalo/packr/v2/packr2@latest
packr2
