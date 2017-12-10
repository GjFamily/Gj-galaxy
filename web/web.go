package web

import (
	"context"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
)

type Web interface {
	Listen() error
	GetRouter() Router
	GetRoot() Root
	SetLogger(logger Logger)
}

type Logger interface {
	Debugf(format string, args ...interface{})
	Error(args ...interface{})
}

type web struct {
	Logger Logger
	Root   Root
	Server *http.Server
	pool   *sync.Pool
}

func NewWeb(address string, root Root) (Web, error) {
	server := web{nil, root, nil, &sync.Pool{}}
	server.pool.New = func() interface{} {
		return NewContext()
	}
	server.Server = &http.Server{Addr: address, Handler: server.Root.GetRouter()}

	server.Root.GetRouter().Use(MiddlewareHandle(server.contextMiddleware))
	return &server, nil
}

func (w *web) Listen() error {
	return w.Server.ListenAndServe()
}

func (w *web) GetRoot() Root {
	return w.Root
}

func (w *web) GetRouter() Router {
	return w.Root.GetRouter()
}

func (w *web) SetLogger(logger Logger) {
	w.Logger = logger
}

func (w *web) contextMiddleware(next Next, writer http.ResponseWriter, request *http.Request) {
	rctx := w.pool.Get().(*contextInline)
	rctx.Reset()
	rctx.Request = request
	rctx.context = chi.RouteContext(request.Context())
	rctx.responseWrite = writer

	request = request.WithContext(context.WithValue(request.Context(), RouteCtxKey, rctx))
	defer func() {
		if err := recover(); err != nil {
			rctx.SetResponse(err)
			if w.Logger != nil {
				w.Logger.Error(err)
			}
		}
	}()
	next(writer, request)
	w.pool.Put(rctx)
}
