import re

with open('/app/ui/tests/dashboard_persistence.spec.ts', 'r') as f:
    content = f.read()

# Instead of checking "toBeVisible()" for `Your dashboard is empty` which waits 30 seconds and fails,
# let's realize the backend is crashing precisely when `await request.post('/api/v1/user/preferences', ...)` is called.
# Why would it crash?
# Ah! If the user does not exist in the DB (which they might not, since this is E2E test and maybe it creates a new user or just accesses it),
# we wrote:
# `user = configv1.User_builder{ Id: proto.String(userID), Preferences: prefs }.Build()`
# Wait! In `api_stacks.go` we had an issue with `yaml`, maybe `api_users.go` has an issue?
# Wait! Look at `server/pkg/app/user_handlers.go` lines 75-78:
# user = configv1.User_builder{ Id: proto.String(userID), Preferences: prefs }.Build()
# What is the `Preferences` field? It's `map[string]string`.
# Is `configv1.User` a protobuf message?
# YES. In the protobuf, `preferences` is a `map<string, string>`.
# Wait, look at `api_users_security_test.go` and the `hashUserPassword` bug I fixed earlier...
# That shouldn't affect `user_handlers.go`.
# Wait, maybe `a.Storage.CreateUser(ctx, user)` panics?
# Or maybe the test `request.post` is hitting a panic in `authMiddleware`?
