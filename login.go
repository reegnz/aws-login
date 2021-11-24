package login

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/config"
)

var (
	defaultFederationURL = "https://signin.aws.amazon.com/federation"
	defaultConsoleURL    = "https://console.aws.amazon.com/"
)

type AWSLogin struct {
	ctx           context.Context
	FederationURL string
	ConsoleURL    string
}

func NewAWSLogin(ctx context.Context) *AWSLogin {
	return &AWSLogin{
		ctx:           ctx,
		FederationURL: defaultFederationURL,
		ConsoleURL:    defaultConsoleURL,
	}
}

func (l *AWSLogin) LoginURL() (string, error) {
	sess, err := l.newSigninSession()
	if err != nil {
		return "", err
	}
	token, err := l.getSigninToken(sess)
	if err != nil {
		return "", err
	}
	req, err := l.prepareLoginRequest(token)
	if err != nil {
		return "", err
	}
	return req, nil
}

func (l *AWSLogin) newSigninSession() (string, error) {
	cfg, err := config.LoadDefaultConfig(l.ctx)
	if err != nil {
		return "", err
	}
	creds, err := cfg.Credentials.Retrieve(l.ctx)
	if err != nil {
		return "", err
	}
	sess, err := json.Marshal(map[string]string{
		"sessionId":    creds.AccessKeyID,
		"sessionKey":   creds.SecretAccessKey,
		"sessionToken": creds.SessionToken,
	})
	if err != nil {
		return "", err
	}
	return string(sess), nil
}

func (l *AWSLogin) getSigninToken(sess string) (string, error) {
	tokenReq, err := l.prepareTokenRequest(sess)
	if err != nil {
		return "", err
	}
	res, err := l.get(tokenReq)
	if err != nil {
		return "", err
	}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var token map[string]interface{}
	err = json.Unmarshal(bodyBytes, &token)
	if err != nil {
		return "", err
	}

	return token["SigninToken"].(string), nil
}

func (l *AWSLogin) get(url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(l.ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	client := http.DefaultClient
	return client.Do(req)
}

func (l *AWSLogin) prepareTokenRequest(sess string) (string, error) {
	params := map[string]string{
		"Action":      "getSigninToken",
		"SessionType": "json",
		"Session":     sess,
	}
	return prepareRequest(l.FederationURL, params)
}

func (l *AWSLogin) prepareLoginRequest(token string) (string, error) {
	params := map[string]string{
		"Action":      "login",
		"Destination": l.ConsoleURL,
		"SigninToken": token,
	}
	return prepareRequest(l.FederationURL, params)
}

func prepareRequest(requestUrl string, params map[string]string) (string, error) {
	req, err := url.Parse(requestUrl)
	if err != nil {
		return "", err
	}
	param := req.Query()
	for k, v := range params {
		param.Set(k, v)
	}
	req.RawQuery = param.Encode()
	return req.String(), nil
}
