package restful

type Response struct {
	Status  int         `json:"status"`
	Result  interface{} `json:"result"`
	Message string      `json:"message"`
}

type PageResult struct {
	Result []interface{} `json:"result"`
	Page   int           `json:"page"`
	Size   int           `json:"size"`
}

func Success(result interface{}) Response {
	return Response{0, result, ""}
}

func PageSuccess(result []interface{}, page int, size int) Response {
	return Response{0, PageResult{result, page, size}, ""}
}
