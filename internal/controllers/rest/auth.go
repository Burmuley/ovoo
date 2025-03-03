package rest

// func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		authHeader := r.Header.Get("Authorization")
// 		if len(authHeader) == 0 {
// 			slog.Error("getting authorization info", "details", "missing authorization header")
// 			writeResponse(w, nil, fmt.Errorf("%w: missing authorization header", entities.ErrValidation), -1)
// 		}

// 		authType := strings.Split(authHeader, " ")[0]
// 		// if

// 	}
// }

// func withAuth(ctx context.Context, h http.Handler, logger *slog.Logger) http.Handler {
// 	authFn := func(rw http.ResponseWriter, req *http.Request) {
// 		start := time.Now()
// 		lrw := loggingResponseWriter{ResponseWriter: rw, status: http.StatusOK, size: 0}
// 		h.ServeHTTP(&lrw, req)
// 		duration := time.Since(start)
// 		logLevel := slog.LevelInfo
// 		if lrw.status < http.StatusOK || lrw.status > 399 {
// 			logLevel = slog.LevelError
// 		}

// 		logger.Log(ctx, logLevel, req.RequestURI,
// 			"method", fmt.Sprintf("%s", req.Method),
// 			"status", lrw.status,
// 			"size", lrw.size,
// 			"duration", duration,
// 		)
// 	}

// 	return http.HandlerFunc(loggingFn)
// }
