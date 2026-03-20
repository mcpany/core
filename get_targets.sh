#!/bin/bash
cd server
go tool cover -func=coverage.out | awk '{if ($3 < 60.0 && $3 != "") print $0}' | sort -n -k 3
