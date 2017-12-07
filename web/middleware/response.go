package middleware

import (
	"Gj-galaxy/web"
	"encoding/json"
	"net/http"
)

func JSONMessageMiddleware(next web.Next, writer http.ResponseWriter, request *http.Request) {
	next(writer, request)
	context := web.RouteContext(request.Context())
	if err := json.NewEncoder(writer).Encode(context.Response); err != nil {
		panic(err)
	}
}
