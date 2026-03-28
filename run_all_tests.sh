cd ui
export CI=true
npx playwright test --project=chromium > test_results.log 2>&1 &
