package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHomePage(t *testing.T) {
	baseURL := "http://localhost:9090"

	var (
		resp *http.Response
		err  error
	)
	resp, err = http.Get(baseURL + "/")

	assert.NoError(t, err, "有错误")
	assert.Equal(t, 200, resp.StatusCode, "应返回状态码 200")
}
