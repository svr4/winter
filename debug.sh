#!/bin/bash

data="$(ps -eo pid,comm | grep -i '.*winter')"

IFS=' '
read -ra ARR <<< $data

if [ -z ${ARR[0]} ] && [ -z ${ARR[1]} ]; then
	echo "Editor is not running."
else
	dlv attach ${ARR[0]} $GOPATH/bin/winter --headless --listen=:62345
fi
