#!/bin/bash
cd server
go run cmd/server/main.go &
sleep 5
cd ../ui
npm run dev &
sleep 5
