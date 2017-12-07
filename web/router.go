package web

import (
	"net/http"

	"github.com/go-chi/chi"
)

type Next func(writer http.ResponseWriter, request *http.Request)
type MiddlewareHandler func(next Next, w http.ResponseWriter, r *http.Request)
type ControllerHandler func(context Context) (interface{}, error)

type RouterNode struct {
	Node *chi.Mux
}

type Router chi.Router

func NewRouter() Router {
	router := RouterNode{}
	router.Node = chi.NewRouter()
	return router.Node
}

func MiddlewareHandle(middle MiddlewareHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		h := func(writer http.ResponseWriter, request *http.Request) {
			next.ServeHTTP(writer, request)
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middle(h, w, r)
		})
	}
}

func ControllerHandle(controller ControllerHandler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		context := RouteContext(request.Context())
		defer func() {
			if err := recover(); err != nil {
				context.SetResponse(err)
			}
		}()
		response, err := controller(context)
		if err != nil {
			context.SetResponse(err)
		} else if response != nil {
			context.SetResponse(response)
		}

	}
}
