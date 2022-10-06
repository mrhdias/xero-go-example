//
// Xero Go Example - oauth-pkce
// Author: Henrique Dias
// Last Modification: 2022-10-06 22:15:29
//
// References:
// Code borrowed from this repository
// https://github.com/nirasan/go-oauth-pkce-code-verifier
//

package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

type CodeVerifier struct {
	Value string
}

const (
	DefaultLength = 32
	MinLength     = 32
	MaxLength     = 96
)

func CreateCodeVerifier() (*CodeVerifier, error) {
	return CreateCodeVerifierWithLength(DefaultLength)
}

func CreateCodeVerifierWithLength(length int) (*CodeVerifier, error) {
	if length < MinLength || length > MaxLength {
		return nil, fmt.Errorf("invalid length: %v", length)
	}
	buf, err := randomBytes(length)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %v", err)
	}
	return CreateCodeVerifierFromBytes(buf)
}

func (v *CodeVerifier) String() string {
	return v.Value
}

func (v *CodeVerifier) CodeChallengePlain() string {
	return v.Value
}

func (v *CodeVerifier) CodeChallengeS256() string {
	h := sha256.New()
	h.Write([]byte(v.Value))
	return encode(h.Sum(nil))
}

func encode(msg []byte) string {
	encoded := base64.StdEncoding.EncodeToString(msg)
	encoded = strings.Replace(encoded, "+", "-", -1)
	encoded = strings.Replace(encoded, "/", "_", -1)
	encoded = strings.Replace(encoded, "=", "", -1)
	return encoded
}

func CreateCodeVerifierFromBytes(b []byte) (*CodeVerifier, error) {
	return &CodeVerifier{
		Value: encode(b),
	}, nil
}

// https://tools.ietf.org/html/rfc7636#section-4.1)
func randomBytes(length int) ([]byte, error) {
	const (
		charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
		csLen   = byte(len(charset))
	)
	output := make([]byte, 0, length)
	for {
		buf := make([]byte, length)
		if _, err := io.ReadFull(rand.Reader, buf); err != nil {
			return nil, fmt.Errorf("failed to read random bytes: %v", err)
		}
		for _, b := range buf {
			// Avoid bias by using a value range that's a multiple of 62
			if b < (csLen * 4) {
				output = append(output, charset[b%csLen])

				if len(output) == length {
					return output, nil
				}
			}
		}
	}
}
