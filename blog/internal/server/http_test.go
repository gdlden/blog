package server

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/stretchr/testify/assert"
)

func TestCustomResponseEncoder_Success(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	err := CustomResponseEncoder(recorder, req, map[string]string{"id": "1"})
	assert.NoError(t, err)
	assert.Equal(t, 200, recorder.Code)

	var resp UnifiedResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "success", resp.Message)
	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "1", data["id"])
}

func TestCustomErrorEncoder_BusinessCodeMapping(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	CustomErrorEncoder(recorder, req, errors.BadRequest("PARAM_ERROR", "param error"))
	assert.Equal(t, 400, recorder.Code)

	var resp UnifiedResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 1001, resp.Code)
	assert.Equal(t, "param error", resp.Message)
}
