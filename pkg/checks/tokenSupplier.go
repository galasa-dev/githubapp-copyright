/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package checks

import (
	"crypto/rsa"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

type githubToken struct {
	token   string
	expires time.Time
}

type TokenSupplier interface {
	GetToken(installation int) (string, error)
}

type TokenSupplierImpl struct {
	key     *rsa.PrivateKey
	reissue time.Time
	// tokenString string

	tokens map[int]githubToken

	githubClient GitHubClient
}

func NewTokenSupplier(githubClient GitHubClient, keyFilePath string) (TokenSupplier, error) {

	var err error = nil
	this := new(TokenSupplierImpl)
	this.githubClient = githubClient

	log.Printf("Using key file %s", keyFilePath)

	var keyBytes []byte
	keyBytes, err = os.ReadFile(keyFilePath)
	if err == nil {
		this.key, err = jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
		if err == nil {

			this.tokens = make(map[int]githubToken)
		}
	}
	return this, err
}

func (this *TokenSupplierImpl) GetToken(installation int) (string, error) {
	var tokenResult string = ""
	var err error = nil

	now := time.Now()

	// check to see if we already have a token
	existingToken, found := this.tokens[installation]
	if found {
		if now.Before(existingToken.expires) {
			return existingToken.token, err
		}
	}

	// as there will only be one installation,  the JWT for the github app will need to be refreshed anyway
	iat := time.Now().Add(-time.Second * 10).UTC()
	exp := time.Now().Add(time.Minute * 10).UTC()
	this.reissue = time.Now().Add(time.Minute * 8)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": 125351,
		"iat": iat.Unix(),
		"exp": exp.Unix(),
	})

	var tokenString string
	tokenString, err = jwtToken.SignedString(this.key)
	if err == nil {

		// We have a valid github jwt token,  now to get the installation token
		accessUrl := fmt.Sprintf("https://api.github.com/app/installations/%v/access_tokens", installation)

		var tokenResponse InstallationToken
		tokenResponse, err = this.githubClient.GetNewToken(accessUrl, tokenString)

		if err == nil {
			var expiresAt time.Time
			expiresAt, err = time.Parse(time.RFC3339, tokenResponse.ExpiresAt)
			if err == nil {

				// take 10 minutes off the expires to make sure we dont get caught at the end of the token life
				expiresAt = expiresAt.Add(-time.Minute * 10)

				newToken := githubToken{
					token:   tokenResponse.Token,
					expires: expiresAt,
				}

				this.tokens[installation] = newToken

				tokenResult = tokenResponse.Token
			}
		}
	}

	return tokenResult, err
}
