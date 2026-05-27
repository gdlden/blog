package server

import (
	"encoding/json"
	stdhttp "net/http"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type UnifiedResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func CustomResponseEncoder(w stdhttp.ResponseWriter, r *stdhttp.Request, v interface{}) error {
	if v == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		return json.NewEncoder(w).Encode(UnifiedResponse{Code: 200, Message: "success"})
	}
	if redirector, ok := v.(http.Redirector); ok {
		url, code := redirector.Redirect()
		stdhttp.Redirect(w, r, url, code)
		return nil
	}
	codec, _ := http.CodecForRequest(r, "Accept")
	data, err := codec.Marshal(UnifiedResponse{Code: 200, Message: "success", Data: v})
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/"+codec.Name())
	_, err = w.Write(data)
	return err
}

func CustomErrorEncoder(w stdhttp.ResponseWriter, r *stdhttp.Request, err error) {
	se := errors.FromError(err)
	businessCode := mapHTTPToBusinessCode(int(se.Code))
	codec, _ := http.CodecForRequest(r, "Accept")
	body, err := codec.Marshal(UnifiedResponse{Code: businessCode, Message: se.Message})
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/"+codec.Name())
	httpStatus := int(se.Code)
	if httpStatus < 100 || httpStatus > 599 {
		httpStatus = 500
	}
	w.WriteHeader(httpStatus)
	w.Write(body)
}

func mapHTTPToBusinessCode(httpCode int) int {
	switch httpCode {
	case 400:
		return 1001
	case 401:
		return 1002
	case 404:
		return 1003
	case 500:
		return 1004
	default:
		return httpCode
	}
}
