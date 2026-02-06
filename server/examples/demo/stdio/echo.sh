#!/bin/bash

read -r input
message=$(echo "$input" | jq -r .message)
echo "{\"message\": \"$message\"}"
