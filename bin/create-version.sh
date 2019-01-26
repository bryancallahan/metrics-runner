#!/bin/bash

# Always work relative to this script...
cd `dirname $0`/..

# Grab version information from repo...
versionShortHash=`git rev-parse --short HEAD`
versionHash=`git rev-parse HEAD`
versionBuildNumber=`git rev-list --all --count $versionHash`

# If we have any changes, append "-dev" so we know the hash is bogus...
if ! `git diff --exit-code > /dev/null` ; then
	versionShortHash="$versionShortHash-dev"
fi

# Write version.json file...
echo "{ \"buildNumber\": $versionBuildNumber, \"shortHash\": \"$versionShortHash\", \"hash\": \"$versionHash\" }" > version.json
