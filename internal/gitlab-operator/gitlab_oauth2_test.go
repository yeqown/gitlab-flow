package gitlabop

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_OAuth2(t *testing.T) {
	v := NewOAuth2Support(&OAuth2Config{
		Host:         "https://git.example.com",
		ServeAddr:    "localhost:2333",
		AccessToken:  "",
		RefreshToken: "",
	})

	s := v.(*gitlabOAuth2Support)

	time.Sleep(5 * time.Second)
	s.triggerAuthorize(context.TODO())
	time.Sleep(10 * time.Second)
}

func Test_OAuth2_authorize(t *testing.T) {
	v := NewOAuth2Support(&OAuth2Config{
		Host:         "https://git.example.com",
		ServeAddr:    "localhost:2333",
		AccessToken:  "",
		RefreshToken: "",
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

func Test_OAuth2_callback(t *testing.T) {
	OAuth2AppID = "gitlab-flow"
	OAuth2AppSecret = "your-secret"

	v := NewOAuth2Support(&OAuth2Config{
		Host:         "https://git.example.com",
		ServeAddr:    "localhost:2333",
		AccessToken:  "",
		RefreshToken: "",
	})

	time.Sleep(30 * time.Second)

	s := v.(*gitlabOAuth2Support)
	req := httptest.NewRequest(http.MethodGet, "http://localhost:2333/callback", nil)
	w := httptest.NewRecorder()
	s.callbackHandl(w, req)

	t.Logf("status: %d", w.Code)
	t.Logf("%v", w.Body.String())
}
