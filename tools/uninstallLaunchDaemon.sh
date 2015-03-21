#!/bin/bash

set -u
set -e

# Make sure we are effectively ROOT
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root (or using sudo)." 1>&2
   exit 1
fi

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
SRC=$DIR/../docs

LAUNCHDIR=/Library/LaunchDaemons
PLIST=local.jackpal.seattlesnowman.plist

cd $LAUNCHDIR
launchctl unload $PLIST
rm $PLIST
