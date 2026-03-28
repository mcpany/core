cd ui
export CI=true
npx playwright test tests/alerts.spec.ts --project=chromium -g "should resolve alert via dropdown and update MTTR" --reporter=html
