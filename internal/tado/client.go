package tado

import (
	"fmt"

	resty "github.com/go-resty/resty/v2"
)

const endpoint_user_info string = "https://my.tado.com/api/v1/me"
const endpoint_user_auth string = "https://auth.tado.com/oauth/token"

type AuthSuccess struct {
	Accesstoken  string `json:"access_token"`
	Tokentype    string `json:"token_type"`
	Refreshtoken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	JTI          string `json:"jti"`
}

func getUserInfo() {
	client := resty.New()
	r, err := client.R().Get(endpoint_user_info)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println(r)
}

/*
client_id=tado-web-app
grant_type=password
scope=home.user
username="vergin@gmx.net"
password="XXX"
client_secret=wZaRN7rpjn3FoNyF5IFuxg9uMzYJcvOoQ8QWiIqS3hfk6gLhVlG57j5YNoZL2Rtc
*/

func authorize() (*AuthSuccess, error) {
	client := resty.New()
	resp, err := client.R().
		SetHeaders(map[string]string{
			"Content-Type":  "application/x-www-form-urlencoded",
			"cache-control": "no-cache",
			"Accept":        "application/json",
		}).
		SetFormData(map[string]string{
			"client_id":     "tado-web-app",
			"grant_type":    "password",
			"scope":         "home.user",
			"username":      "vergin@gmx.net",
			"password":      "xxx",
			"client_secret": "wZaRN7rpjn3FoNyF5IFuxg9uMzYJcvOoQ8QWiIqS3hfk6gLhVlG57j5YNoZL2Rtc",
		}).
		SetResult(&AuthSuccess{}).
		Post(endpoint_user_auth)
	if err != nil {
		return nil, err
	}
	return resp.Result().(*AuthSuccess), nil

}
