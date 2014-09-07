#!/bin/bash


PIPE="/tmp/test-pipe"

if [ ! -e "$PIPE" ]; then
	echo "Failure no pipe found at $PIPE"
	exit 1
fi

blast() {
	id=${1:-"unk"}
	dt=$(date +%s)

	if [ ! -e "$PIPE" ]; then
		break
	fi

	echo "This $id is a message $dt" > "$PIPE"
}


for i in $(seq 1 15); do
	if [ ! -e "$PIPE" ]; then
		break
	fi

	for q in $(seq 1 20); do
		$(blast $q)
	done
	sleep 1
done

echo "Done"
