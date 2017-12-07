package restful

import (
	"Gj-galaxy/web"
	"net/http"
)

func throwError(status int, message string) Response {
	return Response{status, nil, message}
}

func ParamError(message string) Response {
	return throwError(400, message)
}

func SystemError(message string) Response {
	return throwError(500, message)
}

func errorHandlerMiddleware(next web.Next, writer http.ResponseWriter, request *http.Request) {
	context := web.RouteContext(request.Context())
	defer func() {
		if err := recover(); err != nil {
			context.SetResponse(SystemError(err.(error).Error()))
			logger.Error(err)
		}
	}()
	next(writer, request)
}
