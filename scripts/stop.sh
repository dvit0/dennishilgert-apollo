#!/bin/bash

socketPath=$1

sudo curl --unix-socket ${socketPath} -i \
  -X PUT "http://localhost/actions" \
  -H "accept: application/json" \
  -H "Content-Type: application/json" \
  -d "{
    \"action_type\": \"SendCtrlAltDel\"
  }"