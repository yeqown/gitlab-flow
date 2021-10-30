package gitlabop

import (
	"context"
	"testing"
	"time"
)

func Test_OAuth2(t *testing.T) {
	v := NewOAuth2Support(&OAuth2Config{
		AppID:            "",
		AppSecret:        "",
		Host:             "https://git.example.com",
		ServeAddr:        "localhost:2333",
		AccessToken:      "",
		RefreshToken:     "",
		RequestTokenHook: nil,
	})

	s := v.(*gitlabOAuth2Support)

	time.Sleep(5 * time.Second)
	s.triggerAuthorize(context.TODO())
	time.Sleep(10 * time.Second)
}

func Test_OAuth2_authorize(t *testing.T) {
	v := NewOAuth2Support(&OAuth2Config{
		AppID:            "",
		AppSecret:        "",
		Host:             "https://git.example.com",
		ServeAddr:        "localhost:2333",
		AccessToken:      "",
		RefreshToken:     "",
		RequestTokenHook: nil,
	})

	s := v.(*gitlabOAuth2Support)

	err := s.requestToken(
		context.TODO(),
		"6968014a4ad19d2640f462110b96bb910d172cf221b9ad7f006bee7808fc8828",
		false,
	)
	if err != nil {
		t.Error(err)
	}
}
