package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIResponseSuccess(t *testing.T) {
	success := &ApiResponse{Status: "success"}
	assert.Nil(t, validateAPIResponse(success, ""))
}

func TestAPIResponseFailed(t *testing.T) {
	success := &ApiResponse{Status: "meh"}
	assert.NotNil(t, validateAPIResponse(success, ""))
}

func TestAPIResponseSuccessWithMessage(t *testing.T) {
	success := &ApiResponse{Status: "success", Message: "meh"}
	assert.Nil(t, validateAPIResponse(success, ""))
}
