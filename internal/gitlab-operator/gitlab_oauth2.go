package gitlabop

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/yeqown/log"

	"github.com/yeqown/gitlab-flow/pkg"
)

const (
	step1URI = "/oauth/authorize"
	step2URI = "/oauth/token"
	// defaultScope = "read_user api read_repository read_registry"
	defaultScope = "read_user api read_repository"
)

var (
	// OAuth2AppID and OAuth2AppSecret
	// DONE(@yeqown) These two parameters should be passed from build parameters to keep application safety.
	// go build -ldflags="-X package/sub.OAuth2AppID=XXX -X package/sub.OAuth2AppSecret=XXX" ./cmd/gitlab-flow
	OAuth2AppID     string
	OAuth2AppSecret string

	errNilOAuth2Application = errors.New(
		"empty app id or secret, visit: https://github.com/yeqown/gitlab-flow#access-token for more detail")
)

// OAuth2Config helps construct gitlab OAuth2 support.
type OAuth2Config struct {
	// Host of gitlab code repository web application, such as: https://git.example.com
	Host string

	// ServeAddr indicates which port will gitlabOAuth2Support will listen
	// to receive callback request from gitlab oauth server.
	ServeAddr string

	// AccessToken, RefreshToken represent tokens stored before,
	// if they are empty, means authorization is needed.
	AccessToken, RefreshToken string

	// // RequestTokenHook will be called while gitlabOAuth2Support get AccessToken and RefreshToken,
	// // but if authorization failed in any step, callback will miss.
	// RequestTokenHook func(accessToken, refreshToken string)
}

func fixOAuthConfig(c *OAuth2Config) error {
	if c == nil {
		return errors.New("empty config")
	}

	if c.ServeAddr == "" {
		c.ServeAddr = "localhost:2333"
	}

	if OAuth2AppID == "" || OAuth2AppSecret == "" {
		return errNilOAuth2Application
	}

	return nil
}

type gitlabOAuth2Support struct {
	oc *OAuth2Config

	// hc represents a http.Client.
	hc *http.Client

	// state is client unique identifier for each oauth authorization.
	state string

	// tokenC would be triggered while new tokens requested.
	tokenC chan struct{}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewOAuth2Support(c *OAuth2Config) IGitlabOauth2Support {
	if err := fixOAuthConfig(c); err != nil {
		panic(err)
	}

	log.
		WithField("config", c).
		Debug("NewOAuth2Support called")
	g := &gitlabOAuth2Support{
		oc: c,
		hc: &http.Client{
			Timeout: 5 * time.Second,
		},
		state:  "todo",
		tokenC: make(chan struct{}),
	}

	go g.serve()

	return g
}

// Enter oauth2 support logic, authorize getting token while RefreshToken is empty,
// but refresh token while RefreshToken is not empty and not expired.
func (g *gitlabOAuth2Support) Enter(refreshToken string) (err error) {
	if refreshToken == "" {
		g.triggerAuthorize(context.TODO())
		return
	}

	// refresh or re-authorize
	err = g.requestToken(context.TODO(), refreshToken, true)
	if err != nil {
		if errors.Is(err, errRefreshTokenExpired) {
			g.triggerAuthorize(context.TODO())
			return nil
		}

		return
	}

	return
}

func (g *gitlabOAuth2Support) Load() (accessToken, refreshToken string) {
	// waits for tokenC channel's signal.
	_, ok := <-g.tokenC
	if !ok {
		// closed
	}

	if g.oc == nil {
		return
	}

	return g.oc.AccessToken, g.oc.RefreshToken
}

//go:embed callback.tmpl
var callbackTmpl embed.FS

func (g *gitlabOAuth2Support) callbackHandl(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	code := r.Form.Get("code")
	state := r.Form.Get("state")
	_error := r.Form.Get("_error")
	errorDescription := r.Form.Get("error_description")

	log.WithFields(log.Fields{
		"code":             code,
		"state":            state,
		"_error":           _error,
		"errorDescription": errorDescription,
	}).Debug("gitlabOAuth2Support serve gets a callback request")

	tmpl, err := template.ParseFS(callbackTmpl, "callback.tmpl")
	if err != nil {
		log.Errorf("gitlabOAuth2Support.callbackHandl parse FS failed: %v", err)
	}

	var (
		status = http.StatusOK
		data   = struct {
			Error        bool
			ErrorMessage string
			Now          string
		}{
			Error:        false,
			ErrorMessage: "",
			Now:          time.Now().Format(time.RFC1123),
		}
	)

	// error check and parameters validation check.
	if len(_error) != 0 {
		data.Error = true
		data.ErrorMessage = fmt.Sprintf("%s (%s)", _error, errorDescription)
		status = http.StatusInternalServerError
		goto render
	}
	if code == "" || state == "" {
		data.Error = true
		data.ErrorMessage = fmt.Sprintf("Invalid Parameter(code:%s empty or state:%s is empty)", code, state)
		status = http.StatusBadRequest
		goto render
	}

	// authorization callback is in line with forecast.
	if err := g.requestToken(r.Context(), code, false); err != nil {
		log.Errorf("gitlabOAuth2Support callbackHandl failed to requestToken: %v", err)
		data.Error = true
		data.ErrorMessage = fmt.Sprintf("Request Token Failed (%s)", err.Error())
		status = http.StatusInternalServerError
		goto render
	}

	log.Info("gitlab-flow oauth authorization succeeded!")

render:
	w.WriteHeader(status)
	if err := tmpl.Execute(w, &data); err != nil {
		log.Errorf("gitlabOAuth2Support.callbackHandl failed to render: %v", err)
	}
}

// serve is serving a backend HTTP server process to
// receive redirect requests from gitlab.
func (g *gitlabOAuth2Support) serve() {
	http.HandleFunc("/callback", g.callbackHandl)
	err := http.ListenAndServe(g.oc.ServeAddr, nil)
	if err != nil {
		log.Errorf("gitlabOAuth2Support serve quit: %v", err)
	}
}

func (g *gitlabOAuth2Support) generateState() string {
	// DONE(@yeqown): replace state calculation with random int
	g.state = strconv.Itoa(rand.Intn(int(time.Now().UnixNano())))
	return g.state
}

func (g *gitlabOAuth2Support) triggerAuthorize(ctx context.Context) {
	form := url.Values{}

	form.Add("client_id", OAuth2AppID)
	form.Add("redirect_uri", fmt.Sprintf("http://%s/callback", g.oc.ServeAddr))
	form.Add("response_type", "code")
	form.Add("state", g.generateState())
	form.Add("scope", defaultScope)

	fmt.Println("Your access token is invalid or expired, please click following link to authorize:")
	uri := fmt.Sprintf("%s%s?%s", g.oc.Host, step1URI, form.Encode())
	fmt.Println(uri)
	if err := pkg.OpenBrowser(uri); err != nil {
		log.
			WithFields(log.Fields{"error": err, "uri": uri}).
			Error("could not open browser")
	}
}

var (
	errRefreshTokenExpired = errors.New("Enter token expired")
)

// requestToken request token from gitlab oauth server with authorization code or refresh_token.
// in case of Enter token expired, should be forced to re-request triggerAuthorize from user.
func (g *gitlabOAuth2Support) requestToken(ctx context.Context, credential string, isRefresh bool) error {
	form := url.Values{}
	form.Add("client_id", OAuth2AppID)
	form.Add("client_secret", OAuth2AppSecret)
	form.Add("redirect_uri", fmt.Sprintf("http://%s/callback", g.oc.ServeAddr))

	switch isRefresh {
	case true:
		form.Add("refresh_token", credential)
		form.Add("grant_type", "refresh_token")
	default:
		form.Add("code", credential)
		form.Add("grant_type", "authorization_code")
	}

	resp := struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		CreatedAt    int64  `json:"created_at"`

		// Error
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}{}

	if err := g._execPost(ctx, step2URI, form, &resp); err != nil {
		return err
	}

	log.
		WithField("response", resp).
		Debug("requestToken response")

	if len(resp.Error) != 0 {
		// got Error
		if resp.Error == "invalid_grant" {
			return errors.Wrap(errRefreshTokenExpired, resp.ErrorDescription)
		}

		close(g.tokenC)
		return fmt.Errorf("gitlab-flow failed request access token: %s: %s", resp.Error, resp.ErrorDescription)
	}

	g.oc.AccessToken = resp.AccessToken
	g.oc.RefreshToken = resp.RefreshToken

	go func() {
		// FIXED: <del> now this operation would be blocked here, since tokenC is a non-buffered channel </del>.
		select {
		case g.tokenC <- struct{}{}:
			close(g.tokenC)
		default:
		}
	}()

	return nil
}

func (g *gitlabOAuth2Support) _execPost(ctx context.Context, uri string, form url.Values, resp interface{}) error {
	uri = fmt.Sprintf("%s%s?%s", g.oc.Host, uri, form.Encode())

	// log.
	//	WithField("uri", uri).
	//	Debug("gitlabOAuth2Support _execPost called")

	r, err := g.hc.Post(uri, "application/json", nil)
	if err != nil {
		return errors.Wrap(err, "failed to _execPost")
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read")
	}

	if err = json.Unmarshal(data, resp); err != nil {
		return errors.Wrap(err, "unmarshal failed")
	}

	return nil
}
