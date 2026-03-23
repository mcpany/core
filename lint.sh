bazelisk run //:lint \
--config=remote \
--remote_header=x-buildbuddy-api-key=vGFjlQg7X49NwoQHAfRW
bazelisk test //ui:lint //ui:typecheck \
--config=remote \
--test_output=errors \
--remote_header=x-buildbuddy-api-key=vGFjlQg7X49NwoQHAfRW
