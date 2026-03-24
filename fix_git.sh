#!/bin/bash
git checkout origin/main -- .github/
git update-index --assume-unchanged .github/scripts/check-results-and-comment.sh .github/workflows/bot-daily.yml .github/workflows/bot-hourly.yml .github/workflows/ci.yml .github/workflows/release.yml
