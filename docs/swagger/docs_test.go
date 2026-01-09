package swagger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSwaggerInfo(t *testing.T) {
	// Verify that SwaggerInfo is populated
	assert.NotNil(t, SwaggerInfo)
	// Title and Version might be empty depending on generation
	assert.Equal(t, "swagger", SwaggerInfo.InfoInstanceName)
	assert.NotEmpty(t, SwaggerInfo.SwaggerTemplate)
}
