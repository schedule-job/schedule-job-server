package oauth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Github struct {
	ClientId             string
	ClientSecret         string
	RedirectUrl          string
	GithubAccessTokenAPI string
	GithubUserAPI        string
	GithubLoginUrl       string
	OAuthInterface
}

type GithubAuthPayload struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	RedirectUrl  string `json:"redirect_uri"`
}

func (g *Github) getAccessToken(code string) (string, error) {
	payload := GithubAuthPayload{
		ClientId:     g.ClientId,
		ClientSecret: g.ClientSecret,
		Code:         code,
		RedirectUrl:  g.RedirectUrl,
	}

	body, marshalErr := json.Marshal(payload)

	if marshalErr != nil {
		return "", marshalErr
	}
	fmt.Println(g.GithubAccessTokenAPI)
	req, reqErr := http.NewRequest("POST", g.GithubAccessTokenAPI, bytes.NewReader(body))

	if reqErr != nil {
		return "", reqErr
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, resErr := client.Do(req)

	if resErr != nil {
		return "", resErr
	}

	defer res.Body.Close()

	var userData map[string]interface{}
	decodeErr := json.NewDecoder(res.Body).Decode(&userData)

	if decodeErr != nil {
		return "", decodeErr
	}

	if userData["error"] != "" && userData["error"] != nil {
		return "", errors.New(userData["error_description"].(string) + " more info : " + userData["error_uri"].(string))
	}

	return userData["access_token"].(string), nil
}

func (g *Github) getUser(accessToken string) (*User, error) {
	req, reqErr := http.NewRequest("GET", g.GithubUserAPI, nil)

	if reqErr != nil {
		return nil, reqErr
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	res, resErr := client.Do(req)

	if resErr != nil {
		return nil, resErr
	}

	read, readErr := io.ReadAll(res.Body)

	if readErr != nil {
		return nil, readErr
	}

	user := User{}

	parseErr := json.Unmarshal(read, &user)

	if parseErr != nil {
		return nil, parseErr
	}

	return &user, nil
}

func (g *Github) GetUser(code string) (*User, error) {
	accessToken, accessTokenErr := g.getAccessToken(code)

	if accessTokenErr != nil {
		return nil, accessTokenErr
	}

	return g.getUser(accessToken)
}

func (g *Github) GetLoginUrl() string {
	return g.GithubLoginUrl + "?scope=user&client_id=" + g.ClientId + "&redirect_url=" + g.RedirectUrl
}
