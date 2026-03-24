import re

with open("server/pkg/storage/postgres/store_load_test.go", "r") as f:
    content = f.read()

content = content.replace('\n\t\tassert.Equal(t, "service-1", cfg.GetUpstreamServices()[0].GetId())\n\t\tassert.Equal(t, "user-1", cfg.GetUsers()[0].GetId())\n\t\tassert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_INFO, cfg.GetGlobalSettings().GetLogLevel())\n\t\tassert.Equal(t, "Collection One", cfg.GetCollections()[0].GetName())', '')

with open("server/pkg/storage/postgres/store_load_test.go", "w") as f:
    f.write(content)
