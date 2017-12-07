package restful

import (
	"net/http"
)

type Router interface {
	Use(...func(http.Handler) http.Handler)
}

type Restful interface {
}

type restful struct {
	router Router
}

func NewRestful(router Router) Restful {
	r := restful{}
	r.router = router

	return &r
}

func (r *restful) Resource(name string) Router {
	//sr := r.router.

	//sr.Use(web.MiddlewareHandle(errorHandlerMiddleware))
	//sr.Use(web.MiddlewareHandle(middleware.JSONMessageMiddleware))

	return nil
}
