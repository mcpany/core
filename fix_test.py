import re

with open("server/pkg/storage/postgres/store_load_test.go", "r") as f:
    content = f.read()

content = content.replace("cfg.GetGlobalSettings().GetMcpListenAddress()", "cfg.GetGlobalSettings().GetLogLevel().String()")
content = content.replace("InstanceId: proto.String(\"instance-1\")", "LogLevel: configv1.GlobalSettings_LOG_LEVEL_INFO.Enum()")
content = content.replace("instance-1", "LOG_LEVEL_INFO")

with open("server/pkg/storage/postgres/store_load_test.go", "w") as f:
    f.write(content)
