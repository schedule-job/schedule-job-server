package oauth

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/schedule-job/schedule-job-server/internal/errorset"
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

	body, errMarshal := json.Marshal(payload)

	if errMarshal != nil {
		log.Fatalln(errMarshal.Error())
		return "", errorset.ErrOAuth
	}

	req, errReq := http.NewRequest("POST", g.GithubAccessTokenAPI, bytes.NewReader(body))

	if errReq != nil {
		log.Fatalln(errReq.Error())
		return "", errorset.ErrOAuth
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, errRes := client.Do(req)

	if errRes != nil {
		log.Fatalln(errRes.Error())
		return "", errorset.ErrOAuth
	}

	defer res.Body.Close()

	var userData map[string]interface{}
	errDecode := json.NewDecoder(res.Body).Decode(&userData)

	if errDecode != nil {
		log.Fatalln(errDecode.Error())
		return "", errorset.ErrOAuth
	}

	if userData["error"] != "" && userData["error"] != nil {
		log.Fatalln(userData["error_description"].(string) + " more info : " + userData["error_uri"].(string))
		return "", errorset.ErrOAuth
	}

	return userData["access_token"].(string), nil
}

func (g *Github) getUser(accessToken string) (*User, error) {
	req, errReq := http.NewRequest("GET", g.GithubUserAPI, nil)

	if errReq != nil {
		log.Fatalln(errReq.Error())
		return nil, errorset.ErrOAuth
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	res, errRes := client.Do(req)

	if errRes != nil {
		log.Fatalln(errRes.Error())
		return nil, errorset.ErrOAuth
	}

	read, errRead := io.ReadAll(res.Body)

	if errRead != nil {
		log.Fatalln(errRead.Error())
		return nil, errorset.ErrOAuth
	}

	user := User{}

	errParse := json.Unmarshal(read, &user)

	if errParse != nil {
		log.Fatalln(errParse.Error())
		return nil, errorset.ErrOAuth
	}

	return &user, nil
}

func (g *Github) GetUser(code string) (*User, error) {
	accessToken, errAccessToken := g.getAccessToken(code)

	if errAccessToken != nil {
		log.Fatalln(errAccessToken.Error())
		return nil, errorset.ErrOAuth
	}

	return g.getUser(accessToken)
}

func (g *Github) GetLoginUrl() string {
	return g.GithubLoginUrl + "?scope=user&client_id=" + g.ClientId + "&redirect_url=" + g.RedirectUrl
}
