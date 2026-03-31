#!/bin/bash
while ! curl -s http://localhost:8080/api/v1/health > /dev/null; do
  sleep 2
done
echo "Server is up!"
