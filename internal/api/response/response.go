package response

type APIResponse struct {
	RequestID string `json:"request_id,omitempty"`
	BizCode   string `json:"biz_code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
}

const (
	CodeOK             = "OK"
	CodeAccepted       = "ACCEPTED"
	CodeInvalidRequest = "INVALID_REQUEST"
	CodeInvalidURL     = "INVALID_URL"
	CodeInvalidTaskID  = "INVALID_TASK_ID"
	CodeTaskNotFound   = "TASK_NOT_FOUND"
	CodeInternalError  = "INTERNAL_ERROR"
)

func Success(code, message string, data any) APIResponse {
	return APIResponse{
		BizCode: code,
		Message: message,
		Data:    data,
	}
}

func Error(code, message string) APIResponse {
	return APIResponse{
		BizCode: code,
		Message: message,
	}
}

func WithRequestID(resp APIResponse, requestID string) APIResponse {
	resp.RequestID = requestID
	return resp
}
