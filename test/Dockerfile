#!/bin/bash
# FROM: ubuntu   # FROM: must be first non-shebang comment

# Test some env var processing
export XXX="a new var"
export BLANK2=

d.env FOO=$XXX
d.env BAR=BAR
d.env XXX          # no =... means just take its current value
d.env NONE
d.env BLANK1=
d.env BLANK2

# Now do some real work
echo HI from the script
cp -r * /tmp/

# asd  # uncomment this line to test error case

d.maintainer dug
d.workdir /tmp
d.expose 8080
d.cmd ls -al /tmp
d.onbuild RUN echo hi from onbuild
