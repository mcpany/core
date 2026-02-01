// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow login endpoint without auth
			if r.URL.Path == "/api/v1/auth/login" {
				next.ServeHTTP(w, r)
				return
			}

			ip := util.GetClientIP(r, trustProxy)
			ctx := util.ContextWithRemoteIP(r.Context(), ip)
			r = r.WithContext(ctx)
			apiKey := a.SettingsManager.GetAPIKey()
			requestKey := r.Header.Get("X-API-Key")
			logging.GetLogger().Info("DEBUG: AuthMiddleware details", "configured_key", apiKey, "request_key", requestKey, "path", r.URL.Path)
			authenticated := false

			// 1. Check Global API Key
			if apiKey != "" {
				requestKey := r.Header.Get("X-API-Key")
				if requestKey == "" {
					requestKey = r.URL.Query().Get("api_key")
				}
				if requestKey == "" {
					authHeader := r.Header.Get("Authorization")
					if strings.HasPrefix(authHeader, "Bearer ") {
						requestKey = strings.TrimPrefix(authHeader, "Bearer ")
					}
				}

				if subtle.ConstantTimeCompare([]byte(requestKey), []byte(apiKey)) == 1 {
					authenticated = true
					// Inject API Key into context if needed
					ctx = auth.ContextWithAPIKey(ctx, requestKey)
					// Global API Key grants Admin privileges (Root Access)
					ctx = auth.ContextWithRoles(ctx, []string{"admin"})
					// Also inject a placeholder user ID so that handlers expecting a user context don't fail
					ctx = auth.ContextWithUser(ctx, "system-admin")
				}
			}

			// 2. Check User Authentication (Basic Auth)
			if !authenticated {
				username, _, ok := r.BasicAuth()
				if ok && a.AuthManager != nil {
					if user, found := a.AuthManager.GetUser(username); found {
						if err := auth.ValidateAuthentication(ctx, user.GetAuthentication(), r); err == nil {
							authenticated = true
							ctx = auth.ContextWithUser(ctx, username)
							if len(user.GetRoles()) > 0 {
								ctx = auth.ContextWithRoles(ctx, user.GetRoles())
							}
						}
					}
				}
			}

			if authenticated {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if !forcePrivateIPOnly && apiKey != "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Sentinel Security: If no API key is configured (and no user auth succeeded), enforce localhost-only access.
			// This prevents accidental exposure of the server to the public internet (RCE risk).
			if apiKey == "" {
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					// Fallback if RemoteAddr is weird, assume host is the string itself
					host = r.RemoteAddr
				}

				// Check if the request is from a loopback address
				ipAddr := net.ParseIP(host)
				if !util.IsPrivateIP(ipAddr) {
					logging.GetLogger().Warn("Blocked public internet request because no API Key is configured", "remote_addr", r.RemoteAddr)
					http.Error(w, "Forbidden: Public access requires an API Key to be configured", http.StatusForbidden)
					return
				}

				// Grant Admin privileges (Root Access) for local development/testing convenience
				// when running in insecure mode (private network, no API key).
				ctx = auth.ContextWithRoles(ctx, []string{"admin"})
				ctx = auth.ContextWithUser(ctx, "system-admin")
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}


// HTTPRequestContextMiddleware injects the HTTP request into the context.
func (a *Application) HTTPRequestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.HTTPRequestContextKey, r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// startGrpcServer starts a gRPC server in a new goroutine. It handles graceful
// shutdown when the context is canceled.
//
// ctx is the context for managing the server's lifecycle.
// wg is a WaitGroup to signal when the server has shut down.
// errChan is a channel for reporting errors during startup.
// name is a descriptive name for the server, used in logging.
// lis is the net.Listener for the server.
// register is a function that registers the gRPC services with the server.
func startGrpcServer(
	ctx context.Context,
	wg *sync.WaitGroup,
	errChan chan<- error,
	readyChan chan<- struct{},
	name string,
	lis net.Listener,
	shutdownTimeout time.Duration,
	server *gogrpc.Server,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		serverLog := logging.GetLogger().With("server", name)

		if server == nil {
			return
		}

		// localCtx is used to signal the shutdown goroutine to exit.
		localCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		shutdownComplete := make(chan struct{})
		go func() {
			defer close(shutdownComplete)
			select {
			case <-ctx.Done():
				// This is the normal shutdown path.
			case <-localCtx.Done():
				// This is the shutdown path for when the server fails to start.
			}

			serverLog.Info("Attempting to gracefully shut down server...")
			stopped := make(chan struct{})
			go func() {
				defer close(stopped)
				server.GracefulStop()
			}()

			timer := time.NewTimer(shutdownTimeout)
			defer timer.Stop()
			select {
			case <-stopped:
				// Successful graceful shutdown.
			case <-timer.C:
				// Graceful shutdown timed out.
				serverLog.Warn("Graceful shutdown timed out, forcing stop.")
				server.Stop()
			}
		}()

		serverLog.Info("gRPC server listening", "port", lis.Addr().String())
		if readyChan != nil {
			readyChan <- struct{}{}
		}
		if err := server.Serve(lis); err != nil && err != gogrpc.ErrServerStopped {
			errChan <- fmt.Errorf("[%s] server failed to serve: %w", name, err)
			cancel() // Signal shutdown goroutine to exit
		}
		<-shutdownComplete
		serverLog.Info("Server shut down.")
	}()
}

// wrapBindError checks if the error is a port conflict and returns a user-friendly error message.
func wrapBindError(err error, serverType, address, flag string) error {
	if strings.Contains(err.Error(), "address already in use") || strings.Contains(err.Error(), "bind: permission denied") {
		return fmt.Errorf("❌ %s server failed to listen on %s: %w\n\n💡 Tip: The port is already in use or restricted. Try using a different port:\n   mcpany run %s <new_port>", serverType, address, err, flag)
	}
	return fmt.Errorf("%s server failed to listen: %w", serverType, err)
}

// startHTTPServer starts an HTTP server in a new goroutine. It handles graceful
// shutdown when the context is canceled.
//
// ctx is the context for managing the server's lifecycle.
// wg is a WaitGroup to signal when the server has shut down.
// errChan is a channel for reporting errors during startup.
// name is a descriptive name for the server, used in logging.
// lis is the net.Listener on which the server will listen.
// handler is the HTTP handler for processing requests.
func startHTTPServer(
	ctx context.Context,
	wg *sync.WaitGroup,
	errChan chan<- error,
	readyChan chan<- struct{},
	name string,
	lis net.Listener,
	handler http.Handler,
	shutdownTimeout time.Duration,
	connState func(net.Conn, http.ConnState),
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		serverLog := logging.GetLogger().With("server", name)
		// Listener passed in is already listening
		if readyChan != nil {
			readyChan <- struct{}{}
		}
		// We don't close lis here because http.Server.Serve closes it,
		// or we might want to close it on shutdown?
		// http.Server.Serve docs: "Serve accepts incoming connections on the Listener l... always returns a non-nil error."
		// It does NOT say it closes the listener.
		// However, Shutdown() closes the listener?
		// "Shutdown gracefully shuts down the server without interrupting any active connections... Shutdown works by first closing all open listeners..."
		// So Server.Shutdown closes it.
		// BUT if we error out before Shutdown, we should close it?
		// Let's rely on Server.Serve or Shutdown closing it, or defer close if not?
		// Ideally we defer close if Serve returns error other than ErrServerClosed.
		// Use a flag to check if Shutdown was called?
		// Or just defer Close() ignoring error? http.Server might close it too, double close is fine for net.Listener usually.
		// But let's check stdlib behavior.
		// Safe bet: defer lis.Close() at the top.
		// If Shutdown is called, it closes it. Double close is harmless for TCP listeners.
		defer func() { _ = lis.Close() }()

		serverLog = serverLog.With("port", lis.Addr().String())

		server := &http.Server{
			Handler: handler,
			BaseContext: func(_ net.Listener) context.Context {
				return ctx
			},
			ConnState: func(c net.Conn, state http.ConnState) {
				if connState != nil {
					connState(c, state)
				}
				switch state {
				case http.StateNew:
					metrics.IncrCounter([]string{"http", "connections", "opened", "total"}, 1)
				case http.StateClosed:
					metrics.IncrCounter([]string{"http", "connections", "closed", "total"}, 1)
				}
			},
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       120 * time.Second,
		}

		// localCtx is used to signal the shutdown goroutine to exit.
		localCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		shutdownComplete := make(chan struct{})
		go func() {
			defer close(shutdownComplete)
			select {
			case <-ctx.Done():
				// This is the normal shutdown path.
			case <-localCtx.Done():
				// This is the shutdown path for when the server fails to start.
			}
			shutdownCtx, cancelTimeout := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancelTimeout()
			serverLog.Info("Attempting to gracefully shut down server...")
			if err := server.Shutdown(shutdownCtx); err != nil {
				serverLog.Error("Shutdown error", "error", err)
			}
		}()

		serverLog.Info("HTTP server listening")
		if err := server.Serve(lis); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("[%s] server failed: %w", name, err)
			cancel() // Signal shutdown goroutine to exit
		}

		<-shutdownComplete
		serverLog.Info("Server shut down.")
	}()
}
