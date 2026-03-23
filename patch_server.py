import sys
content = open('server/pkg/app/server.go').read()
if 'standardMiddlewares.RecursiveContext == nil' not in content:
    # Add Recursive Context registration
    search = 'if standardMiddlewares != nil && standardMiddlewares.Debugger != nil {'
    replace = '''	// Register Recursive Context Manager
	if standardMiddlewares != nil && standardMiddlewares.RecursiveContext == nil {
		standardMiddlewares.RecursiveContext = middleware.NewRecursiveContextManager()
	}
	if standardMiddlewares != nil {
		mux.Handle("/context/session", authMiddleware(standardMiddlewares.RecursiveContext.APIHandler()))
		mux.Handle("/context/session/", authMiddleware(standardMiddlewares.RecursiveContext.APIHandler()))
	}

	''' + search
    content = content.replace(search, replace)

    # Add Recursive Context Middleware
    search2 = 'if standardMiddlewares.Debugger != nil {'
    replace2 = '''// Recursive Context
		if standardMiddlewares.RecursiveContext != nil {
			finalHandler = standardMiddlewares.RecursiveContext.HandleContext(finalHandler)
		}
		''' + search2
    content = content.replace(search2, replace2)

    # Remove non-ASCII
    content = content.replace('❌', '[ERROR]').replace('💡', '[TIP]')

    with open('server/pkg/app/server.go', 'w') as f:
        f.write(content)
