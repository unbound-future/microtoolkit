package model

type ErrorResponse struct {
	Code      int        `json:"code"`
	Msg       string     `json:"msg"`
	ErrorInfo *ErrorInfo `json:"error,omitempty"`
}

type ErrorInfo struct {
	LogID          string `json:"log_id"`
	Troubleshooter string `json:"troubleshooter"`
}

func Error(code int, msg string, logID string, troubleshooter string) *ErrorResponse {
	return &ErrorResponse{
		Code: code,
		Msg:  msg,
		ErrorInfo: &ErrorInfo{
			LogID:          logID,
			Troubleshooter: troubleshooter,
		},
	}
}
