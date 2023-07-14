#!/bin/bash
tmpfile=$(mktemp)

./jsonify_dump.py $(find $DUMP_DIR -name "*.protodump") 2> $tmpfile
cat $tmpfile