cd server
sed -i 's|"github.com/mcpany/core/server/pkgconsts"|"github.com/mcpany/core/server/pkg/consts"|g' cmd/server/main.go
sed -i 's|"github.com/mcpany/core/server/pkgconsts"|"github.com/mcpany/core/server/pkg/consts"|g' pkg/app/api_system.go
