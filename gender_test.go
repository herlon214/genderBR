package genderBR

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGender(t *testing.T) {
	names := []string{"João"}
	results := For(names)
	assert.NotEmpty(t, results)

	assert.Equal(t, "Male", results[0].Gender)
}
