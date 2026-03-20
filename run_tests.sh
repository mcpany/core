cd server
go test -v ./pkg/tool/... -run TestLocalCommandTool_Python_Arg_Space_Issue
go test -v ./pkg/tool/... -run TestLocalCommandTool_Python_Space_Issue
