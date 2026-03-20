cd server
export CI=true
export GITHUB_ACTIONS=true
../build/env/bin/bazelisk test //server/tests/integration:integration_test --test_output=errors
