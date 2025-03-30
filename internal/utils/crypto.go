package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
)

// PubKeyToString converts a public key to a string
func PubKeyToString(pubKey *ecdsa.PublicKey) string {
	return hex.EncodeToString(elliptic.Marshal(pubKey.Curve, pubKey.X, pubKey.Y))
}

// StringToPubKey converts a string to a public key
func StringToPubKey(pubKeyStr string) *ecdsa.PublicKey {
	pubBytes, _ := hex.DecodeString(pubKeyStr)
	x, y := elliptic.Unmarshal(elliptic.P256(), pubBytes)
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
}
