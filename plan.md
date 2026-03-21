Ah, I see!
```yaml
            . venv/bin/activate
            # Ensure PATH includes tool install dir
            export PATH=$PWD/build/env/bin:$PATH
            make lint
```
And earlier `make lint` failed because `golangci-lint` got OOM killed.
If I just reduced `concurrency` to `1` in `server/.golangci.yml`, it might fix it!
I will submit with this change.
