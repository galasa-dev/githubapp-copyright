/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package checks

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	key         *rsa.PrivateKey
	reissue     time.Time
	tokenString string

	tokens map[int]githubToken

	tokenHttpClient http.Client
)

type githubToken struct {
	token   string
	expires time.Time
}

func init() {
	keyBytes, err := os.ReadFile("../../key.pem")
	if err != nil {
		panic(err)
	}

	key, err = jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		panic(err)
	}

	tokens = make(map[int]githubToken)
}

func getToken(installation int) string {

	now := time.Now()

	// check to see if we already have a token
	existingToken, found := tokens[installation]
	if found {
		if now.Before(existingToken.expires) {
			return existingToken.token
		}
	}

	// as there will only be one installation,  the JWT for the github app will need to be refreshed anyway
	iat := time.Now().Add(-time.Second * 10).UTC()
	exp := time.Now().Add(time.Minute * 10).UTC()
	reissue = time.Now().Add(time.Minute * 8)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": 125351,
		"iat": iat.Unix(),
		"exp": exp.Unix(),
	})

	tokenString, err := token.SignedString(key)
	if err != nil {
		panic(err) // TODO
	}

	// We have a valid github jwt token,  now to get the installation token

	accessUrl := fmt.Sprintf("https://api.github.com/app/installations/%v/access_tokens", installation)

	req, err := http.NewRequest("POST", accessUrl, nil)
	if err != nil {
		panic(err) // TODO
	}

	req.Header.Add("Authorization", "Bearer "+tokenString)
	req.Header.Add("Accept", "application/vnd.github.v3+json")

	resp, err := tokenHttpClient.Do(req)
	if err != nil {
		panic(err) // TODO
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		panic(resp.Status) // TODO
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err) // TODO
	}

	var tokenResponse InstallationToken
	err = json.Unmarshal(bodyBytes, &tokenResponse)
	if err != nil {
		panic(err) // TODO
	}

	expiresAt, err := time.Parse(time.RFC3339, tokenResponse.ExpiresAt)
	if err != nil {
		panic(err) // TODO
	}

	// take 10 minutes off the expires to make sure we dont get caught at the end of the token life
	expiresAt = expiresAt.Add(-time.Minute * 10)

	newToken := githubToken{
		token:   tokenResponse.Token,
		expires: expiresAt,
	}

	tokens[installation] = newToken

	return tokenResponse.Token
}
