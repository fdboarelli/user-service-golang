//go:build unit
// +build unit

package utility

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var hashingUtil *Utility

func setupService() *Utility {
	hashingUtil = New("TestSecretKey")
	return hashingUtil
}

func TestHashing(t *testing.T) {
	hashingUtil = setupService()
	inputToBeEncode := "MyPlaintextPassword"
	// run test and validate
	encryptedInput := hashingUtil.EncodeInput(inputToBeEncode)
	assert.Equal(t, "a054183598b2fca6b81324bf6d333955e7cbcfd17331ae65e1f9c2a99b89df37", encryptedInput)
}
