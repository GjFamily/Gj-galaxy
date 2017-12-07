package restful

import (
	"Gj-galaxy/web"
	"Gj-galaxy/web/middleware"
	"net/http"
)

type Router interface {
	GetRouter() Router
	Use(func(http.Handler) http.Handler)
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
	sr := r.router.GetRouter()

	sr.Use(web.MiddlewareHandle(errorHandlerMiddleware))
	sr.Use(web.MiddlewareHandle(middleware.JSONMessageMiddleware))

	return sr
}
