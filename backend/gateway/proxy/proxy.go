package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gin-gonic/gin"
	"github.com/hesham-ashraf/LearnVibe/backend/gateway/config"
	"github.com/sony/gobreaker"
)

// ServiceProxy manages proxying requests to microservices
type ServiceProxy struct {
	CMSUrl         *url.URL
	ContentUrl     *url.URL
	config         *config.Config
	cmsBreaker     *gobreaker.CircuitBreaker
	contentBreaker *gobreaker.CircuitBreaker
}

// NewServiceProxy creates a new service proxy
func NewServiceProxy(cfg *config.Config) (*ServiceProxy, error) {
	cmsUrl, err := url.Parse(cfg.CMSServiceURL)
	if err != nil {
		return nil, fmt.Errorf("invalid CMS service URL: %v", err)
	}

	contentUrl, err := url.Parse(cfg.ContentServiceURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Content service URL: %v", err)
	}

	// Configure circuit breakers for each service
	cmsSetting := gobreaker.Settings{
		Name:        "CMSServiceBreaker",
		MaxRequests: cfg.CircuitBreakerMaxRequests,
		Interval:    time.Duration(cfg.CircuitBreakerInterval) * time.Second,
		Timeout:     time.Duration(cfg.CircuitBreakerTimeout) * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= 0.5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("Circuit breaker %s state changed from %s to %s\n", name, from, to)
		},
	}

	contentSetting := gobreaker.Settings{
		Name:        "ContentServiceBreaker",
		MaxRequests: cfg.CircuitBreakerMaxRequests,
		Interval:    time.Duration(cfg.CircuitBreakerInterval) * time.Second,
		Timeout:     time.Duration(cfg.CircuitBreakerTimeout) * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= 0.5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("Circuit breaker %s state changed from %s to %s\n", name, from, to)
		},
	}

	return &ServiceProxy{
		CMSUrl:         cmsUrl,
		ContentUrl:     contentUrl,
		config:         cfg,
		cmsBreaker:     gobreaker.NewCircuitBreaker(cmsSetting),
		contentBreaker: gobreaker.NewCircuitBreaker(contentSetting),
	}, nil
}

// ProxyCMSRequest proxies requests to the CMS service
func (sp *ServiceProxy) ProxyCMSRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Execute request through circuit breaker
		result, err := sp.cmsBreaker.Execute(func() (interface{}, error) {
			// Create a reverse proxy
			proxy := httputil.NewSingleHostReverseProxy(sp.CMSUrl)

			// Set default request director
			originalDirector := proxy.Director
			proxy.Director = func(req *http.Request) {
				originalDirector(req)
				sp.modifyRequest(req, sp.CMSUrl)
			}

			// Handle errors
			proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
				fmt.Printf("CMS proxy error: %v\n", err)
				rw.WriteHeader(http.StatusBadGateway)
				rw.Write([]byte("CMS service unavailable"))
			}

			// Serve the request with retry logic
			var proxyErr error
			operation := func() error {
				ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(sp.config.RequestTimeout)*time.Second)
				defer cancel()

				req := c.Request.WithContext(ctx)
				proxy.ServeHTTP(c.Writer, req)

				// If we've written a bad gateway status, return an error to trigger retry
				if c.Writer.Status() == http.StatusBadGateway {
					return fmt.Errorf("received bad gateway status")
				}

				return nil
			}

			backoffConfig := backoff.NewExponentialBackOff()
			backoffConfig.MaxElapsedTime = 5 * time.Second

			if err := backoff.Retry(operation, backoffConfig); err != nil {
				proxyErr = err
			}

			if proxyErr != nil {
				return nil, proxyErr
			}

			return nil, nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "CMS service is currently unavailable",
			})
			return
		}

		_ = result // result is not used as response is written by the proxy
		c.Abort()  // Prevent further handlers from executing
	}
}

// ProxyContentRequest proxies requests to the Content service
func (sp *ServiceProxy) ProxyContentRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Execute request through circuit breaker
		result, err := sp.contentBreaker.Execute(func() (interface{}, error) {
			// Create a reverse proxy
			proxy := httputil.NewSingleHostReverseProxy(sp.ContentUrl)

			// Set default request director
			originalDirector := proxy.Director
			proxy.Director = func(req *http.Request) {
				originalDirector(req)
				sp.modifyRequest(req, sp.ContentUrl)
			}

			// Handle errors
			proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
				fmt.Printf("Content proxy error: %v\n", err)
				rw.WriteHeader(http.StatusBadGateway)
				rw.Write([]byte("Content service unavailable"))
			}

			// Serve the request with retry logic
			var proxyErr error
			operation := func() error {
				ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(sp.config.RequestTimeout)*time.Second)
				defer cancel()

				req := c.Request.WithContext(ctx)
				proxy.ServeHTTP(c.Writer, req)

				// If we've written a bad gateway status, return an error to trigger retry
				if c.Writer.Status() == http.StatusBadGateway {
					return fmt.Errorf("received bad gateway status")
				}

				return nil
			}

			backoffConfig := backoff.NewExponentialBackOff()
			backoffConfig.MaxElapsedTime = 5 * time.Second

			if err := backoff.Retry(operation, backoffConfig); err != nil {
				proxyErr = err
			}

			if proxyErr != nil {
				return nil, proxyErr
			}

			return nil, nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "Content service is currently unavailable",
			})
			return
		}

		_ = result // result is not used as response is written by the proxy
		c.Abort()  // Prevent further handlers from executing
	}
}

// modifyRequest modifies the incoming request for the target service
func (sp *ServiceProxy) modifyRequest(req *http.Request, target *url.URL) {
	targetQuery := target.RawQuery
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)

	// Set the appropriate Host header
	req.Host = target.Host

	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
}

// joinURLPath joins base and reference URL paths
func joinURLPath(base, ref *url.URL) (string, string) {
	if ref.Path == "" {
		return base.Path, base.RawPath
	}

	path := base.Path
	if path == "" || path == "/" {
		return ref.Path, ref.RawPath
	}

	// If base path doesn't end with / and ref path doesn't start with /
	// Add / between them
	if path[len(path)-1] != '/' && ref.Path[0] != '/' {
		return path + "/" + ref.Path, ""
	}

	// If base path ends with / and ref path starts with /
	// Remove one of the slashes
	if path[len(path)-1] == '/' && ref.Path[0] == '/' {
		return path + ref.Path[1:], ""
	}

	// Otherwise just concatenate them
	return path + ref.Path, ""
}
