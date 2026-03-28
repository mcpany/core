cd ui
export CI=false
npm run dev &
sleep 5
npx playwright test tests/alerts.spec.ts --project=chromium -g "should resolve alert via dropdown and update MTTR"
