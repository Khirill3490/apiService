package api

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func NewErrorResponse(err error) Response {
	return Response{
		Status: "error",
		Error:  err.Error(),
	}
}

func NewSuccessResponse() Response {
	return Response{
		Status: "success",
	}
}
