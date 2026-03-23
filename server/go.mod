module github.com/mcpany/core/server

go 1.26.1

toolchain go1.26.1

replace github.com/mcpany/core => ../
replace github.com/mcpany/core/proto => ../proto

require (
	al.essio.dev/pkg/shellescape v1.6.0
	cloud.google.com/go/storage v1.58.0
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/JohannesKaufmann/html-to-markdown v1.6.0
	github.com/Masterminds/semver/v3 v3.4.0
	github.com/PaesslerAG/jsonpath v0.1.1
	github.com/alexliesenfeld/health v0.8.1
	github.com/alicebob/miniredis/v2 v2.35.0
	github.com/alitto/pond/v2 v2.5.0
	github.com/antchfx/xmlquery v1.4.4
	github.com/antchfx/xpath v1.3.5
)
