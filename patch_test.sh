cd ui
export PATH="$PATH:$HOME/.npm-global/bin"
# In Vite tests with Playwright we need the dev server to start
# But it fails because the backend is not running at 50050.
# The `playwright.config.ts` might define the webServer setup.
