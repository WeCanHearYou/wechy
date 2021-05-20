package web

import (
	"context"
	"fmt"
	"io/ioutil"
	stdLog "log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"time"

	"github.com/getfider/fider/app/models/dto"
	"github.com/getfider/fider/app/pkg/env"
	"github.com/getfider/fider/app/pkg/errors"
	"github.com/getfider/fider/app/pkg/log"
	"github.com/getfider/fider/app/pkg/rand"
	"github.com/getfider/fider/app/pkg/worker"
	"github.com/julienschmidt/httprouter"
	cache "github.com/patrickmn/go-cache"
)

// fonts.gstatic.com and fonts.googleapis.com are required for Custom CSS to work with Google Fonts
// unsafe-inline on Style is required for rendering Tags on SSR

var (
	cspBase    = "base-uri 'self'"
	cspDefault = "default-src 'self'"
	cspStyle   = "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com %[2]s"
	cspScript  = "script-src 'self' 'nonce-%[1]s' https://www.google-analytics.com %[2]s"
	cspFont    = "font-src 'self' https://fonts.gstatic.com data: %[2]s"
	cspImage   = "img-src 'self' https: data: %[2]s"
	cspObject  = "object-src 'none'"
	cspFrame   = "frame-src 'self'"
	cspMedia   = "media-src 'none'"
	cspConnect = "connect-src 'self' https://www.google-analytics.com %[2]s"

	//CspPolicyTemplate is the template used to generate the policy
	CspPolicyTemplate = fmt.Sprintf("%s; %s; %s; %s; %s; %s; %s; %s; %s; %s", cspBase, cspDefault, cspStyle, cspScript, cspImage, cspFont, cspObject, cspMedia, cspConnect, cspFrame)
)

type notFoundHandler struct {
	engine  *Engine
	handler HandlerFunc
}

func (h *notFoundHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ctx := NewContext(h.engine, req, res, nil)
	_ = h.handler(ctx)
}

//HandlerFunc represents an HTTP handler
type HandlerFunc func(*Context) error

//MiddlewareFunc represents an HTTP middleware
type MiddlewareFunc func(HandlerFunc) HandlerFunc

//Engine is our web engine wrapper
type Engine struct {
	context.Context
	mux         *httprouter.Router
	renderer    *Renderer
	binder      *DefaultBinder
	middlewares []MiddlewareFunc
	worker      worker.Worker
	server      *http.Server
	cache       *cache.Cache
}

//New creates a new Engine
func New() *Engine {
	ctx := context.Background()
	ctx = log.WithProperties(ctx, dto.Props{
		log.PropertyKeyContextID: rand.String(32),
		log.PropertyKeyTag:       "WEB",
	})

	router := &Engine{
		Context:     ctx,
		mux:         httprouter.New(),
		renderer:    NewRenderer(),
		binder:      NewDefaultBinder(),
		middlewares: make([]MiddlewareFunc, 0),
		worker:      worker.New(),
		cache:       cache.New(5*time.Minute, 10*time.Minute),
	}

	return router
}

//Start the server.
func (e *Engine) Start(address string) {
	log.Info(e, "Application is starting")
	log.Infof(e, "Version: @{Version}", dto.Props{
		"Version": env.Version(),
	})
	log.Infof(e, "Environment: @{Env}", dto.Props{
		"Env": env.Config.Environment,
	})

	var (
		certFilePath = ""
		keyFilePath  = ""
	)

	if env.Config.TLS.Certificate != "" {
		certFilePath = env.Etc(env.Config.TLS.Certificate)
		keyFilePath = env.Etc(env.Config.TLS.PrivateKey)
	}

	stdLog.SetOutput(ioutil.Discard)
	e.server = &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Addr:         address,
		Handler:      e.mux,
		TLSConfig:    getDefaultTLSConfig(),
	}

	for i := 0; i < runtime.NumCPU()*2; i++ {
		go e.Worker().Run(strconv.Itoa(i))
	}

	var (
		err         error
		certManager *CertificateManager
	)
	if env.Config.TLS.Automatic {
		certManager, err = NewCertificateManager(certFilePath, keyFilePath)
		if err != nil {
			panic(errors.Wrap(err, "failed to initialize CertificateManager"))
		}

		e.server.TLSConfig.GetCertificate = certManager.GetCertificate
		log.Infof(e, "https (auto ssl) server started on @{Address}", dto.Props{"Address": address})
		go certManager.StartHTTPServer()
		err = e.server.ListenAndServeTLS("", "")
	} else if certFilePath == "" && keyFilePath == "" {
		log.Infof(e, "http server started on @{Address}", dto.Props{"Address": address})
		err = e.server.ListenAndServe()
	} else {
		log.Infof(e, "https server started on @{Address}", dto.Props{"Address": address})
		err = e.server.ListenAndServeTLS(certFilePath, keyFilePath)
	}

	if err != nil && err != http.ErrServerClosed {
		panic(errors.Wrap(err, "failed to start server"))
	}
}

//Stop the server.
func (e *Engine) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Info(e, "server is shutting down")
	if err := e.server.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shutdown server")
	}
	log.Info(e, "server has shutdown")

	log.Info(e, "worker is shutting down")
	if err := e.worker.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shutdown worker")
	}
	log.Info(e, "worker has shutdown")

	return nil
}

//Cache returns current cache
func (e *Engine) Cache() *cache.Cache {
	return e.cache
}

//Worker returns current worker reference
func (e *Engine) Worker() worker.Worker {
	return e.worker
}

//Group creates a new route group
func (e *Engine) Group() *Group {
	g := &Group{
		engine:      e,
		middlewares: e.middlewares,
	}
	return g
}

//Use adds a middleware to the root engine
func (e *Engine) Use(middleware MiddlewareFunc) {
	e.middlewares = append(e.middlewares, middleware)
}

//Get handles HTTP GET requests
func (e *Engine) Get(path string, handler HandlerFunc) {
	e.mux.Handle("GET", path, e.handle(e.middlewares, handler))
}

//Post handles HTTP POST requests
func (e *Engine) Post(path string, handler HandlerFunc) {
	e.mux.Handle("POST", path, e.handle(e.middlewares, handler))
}

//Put handles HTTP PUT requests
func (e *Engine) Put(path string, handler HandlerFunc) {
	e.mux.Handle("PUT", path, e.handle(e.middlewares, handler))
}

//Delete handles HTTP DELETE requests
func (e *Engine) Delete(path string, handler HandlerFunc) {
	e.mux.Handle("DELETE", path, e.handle(e.middlewares, handler))
}

//NotFound register how to handle routes that are not found
func (e *Engine) NotFound(handler HandlerFunc) {
	e.mux.NotFound = &notFoundHandler{
		engine:  e,
		handler: handler,
	}
}

func (e *Engine) handle(middlewares []MiddlewareFunc, handler HandlerFunc) httprouter.Handle {
	next := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		next = middlewares[i](next)
	}
	var h = func(res http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		params := make(StringMap)
		for _, p := range ps {
			params[p.Key] = p.Value
		}
		ctx := NewContext(e, req, res, params)
		_ = next(ctx)
	}
	return h
}

//Group is our router group wrapper
type Group struct {
	engine      *Engine
	middlewares []MiddlewareFunc
}

//Group creates a new route group
func (g *Group) Group() *Group {
	g2 := &Group{
		engine:      g.engine,
		middlewares: g.middlewares,
	}
	return g2
}

//Use adds a middleware to current route stack
func (g *Group) Use(middleware MiddlewareFunc) {
	g.middlewares = append(g.middlewares, middleware)
}

//Get handles HTTP GET requests
func (g *Group) Get(path string, handler HandlerFunc) {
	g.engine.mux.Handle("GET", path, g.engine.handle(g.middlewares, handler))
}

//Post handles HTTP POST requests
func (g *Group) Post(path string, handler HandlerFunc) {
	g.engine.mux.Handle("POST", path, g.engine.handle(g.middlewares, handler))
}

//Put handles HTTP PUT requests
func (g *Group) Put(path string, handler HandlerFunc) {
	g.engine.mux.Handle("PUT", path, g.engine.handle(g.middlewares, handler))
}

//Delete handles HTTP DELETE requests
func (g *Group) Delete(path string, handler HandlerFunc) {
	g.engine.mux.Handle("DELETE", path, g.engine.handle(g.middlewares, handler))
}

// Static return files from given folder
func (g *Group) Static(prefix, root string) {
	fi, err := os.Stat(env.Path(root))
	if err != nil {
		panic(fmt.Sprintf("Path '%s' not found", root))
	}

	var h HandlerFunc
	if fi.IsDir() {
		h = func(c *Context) error {
			filePath := path.Join(root, c.Param("filepath"))
			fi, err := os.Stat(filePath)
			if err == nil && !fi.IsDir() {
				http.ServeFile(c.Response, c.Request.instance, filePath)
				return nil
			}
			return c.NotFound()
		}
	} else {
		h = func(c *Context) error {
			http.ServeFile(c.Response, c.Request.instance, root)
			return nil
		}
	}
	g.engine.mux.Handle("GET", prefix, g.engine.handle(g.middlewares, h))
}

// ParseCookie return a list of cookie parsed from raw Set-Cookie
func ParseCookie(s string) *http.Cookie {
	if s == "" {
		return nil
	}
	return (&http.Response{Header: http.Header{"Set-Cookie": {s}}}).Cookies()[0]
}
