#!/bin/sh
# The repo vet gate — run this, not bare `go vet ./...`.
#
# ONE deliberate exception: proto/v235's Buffer.WriteByte(v byte) does not
# (and must not) conform to io.ByteWriter — see the comment on the method
# (proto/v235/buffer.go). go vet has no inline suppression, so stdmethods
# is disabled for that one package and on for everything else.
set -e
go vet -stdmethods=false ./proto/v235
go vet $(go list ./... | grep -v 'proto/v235$')
echo "vet: OK (stdmethods exception: proto/v235 — see scripts/vet.sh)"
