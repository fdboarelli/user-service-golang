package utility

// This class implements functional logic that can be used across all the service's implementations

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
)

type Utility struct {
	secretKey string
}

func New(secretKey string) *Utility {
	return &Utility{secretKey: secretKey}
}

// EncodeInput returns the resulting Hex string of the Hmac256 hash of the input
func (utility *Utility) EncodeInput(input string) string {
	log.Info("Starting hashing procedure.")
	hash := hmac.New(sha256.New, []byte(utility.secretKey))
	// Write Data to it
	hash.Write([]byte(input))
	// Get result and encode as hexadecimal string
	hexSha := hex.EncodeToString(hash.Sum(nil))
	log.Debug("Computed hash is: ", hexSha)
	log.Info("Completed hashing procedure.")
	return hexSha
}
