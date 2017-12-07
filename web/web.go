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
}

type web struct {
	Router Router
	Server *http.Server
	pool   *sync.Pool
}

func NewWeb(address string, router Router) (Web, error) {
	server := web{router, nil, &sync.Pool{}}
	server.pool.New = func() interface{} {
		return NewContext()
	}
	server.Server = &http.Server{Addr: address, Handler: server.Router}

	server.Router.Use(MiddlewareHandle(server.contextMiddleware))
	return &server, nil
}

func (server *web) Listen() error {
	return server.Server.ListenAndServe()
}

func (server *web) GetRouter() Router {
	return server.Router
}

func (server *web) contextMiddleware(next Next, writer http.ResponseWriter, request *http.Request) {
	rctx := server.pool.Get().(*contextInline)
	rctx.Reset()
	rctx.Request = request
	rctx.context = chi.RouteContext(request.Context())
	rctx.responseWrite = writer

	request = request.WithContext(context.WithValue(request.Context(), RouteCtxKey, rctx))
	defer func() {
		if err := recover(); err != nil {
			rctx.SetResponse(err)
			logger.Error(err)
		}
	}()
	next(writer, request)
	server.pool.Put(rctx)
}
