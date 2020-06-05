package commons

type JsonResponse struct {
	ErrorCode int
	Data interface{}
}

func (jsonResponse *JsonResponse) Setter(code int, Data interface{}) {
	jsonResponse.ErrorCode, jsonResponse.Data = code, Data
}