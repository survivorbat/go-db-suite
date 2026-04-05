package dbsuite

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToAlphanumeric_OutputsExpectedName(t *testing.T) {
	t.Parallel()
	// Arrange
	in := "Hello I'm so happy! 😊 I'm writing this test to ensure no weird DB names come through! That'd be bad 🚨"

	// Act
	actual := toSafeDBName(in)

	// Assert
	assert.Equal(t, "appyimwritingthistesttoensurenoweirddbnamescomethroughthatdbebad", actual)
}
