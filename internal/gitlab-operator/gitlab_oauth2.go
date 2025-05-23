package gitlabop

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	htmltpl "html/template"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	texttpl "text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/yeqown/log"

	"github.com/yeqown/gitlab-flow/internal/types"
	"github.com/yeqown/gitlab-flow/pkg"
)

const (
	step1URI    = "/oauth/authorize"
	step2URI    = "/oauth/token"
	callbackURI = "/callback"
)

var (
	errNilOAuth2Application = errors.New(
		"empty app id or secret, visit: https://github.com/yeqown/gitlab-flow#access-token for more detail")

	// SecretKey is used to encrypt and decrypt access token and refresh token.
	// NOTE: this key must be 8 bytes long.
	SecretKey = "aflowcli"
)

// OAuth2Config helps construct gitlab OAuth2 support.
type OAuth2Config struct {
	// Host of gitlab code repository web application, such as https://git.example.com
	Host string

	// ServeAddr indicates which port will gitlabOAuth2Support will listen
	// to receive callback request from gitlab oauth server.
	ServeAddr string

	// AppID and AppSecret are application id and secret of gitlab oauth2 application.
	AppID, AppSecret string

	// AccessToken, RefreshToken represent tokens stored before,
	// if they are empty, means authorization is needed.
	AccessToken, RefreshToken string

	// Scopes is a string of scopes, such as "api read_user"
	Scopes string

	// Mode represents the mode of OAuth2 authorization.
	Mode types.OAuth2Mode
}

func NewOAuth2ConfigFrom(cfg *types.Config) *OAuth2Config {
	return &OAuth2Config{
		Host:         cfg.GitlabHost,
		ServeAddr:    cfg.OAuth2.CallbackHost,
		AccessToken:  cfg.OAuth2.AccessToken,                                      // empty
		RefreshToken: cfg.OAuth2.RefreshToken,                                     // empty
		AppID:        pkg.MustDesDecrypt(cfg.OAuth2.AppID, []byte(SecretKey)),     //
		AppSecret:    pkg.MustDesDecrypt(cfg.OAuth2.AppSecret, []byte(SecretKey)), //
		Scopes:       cfg.OAuth2.Scopes,
		Mode:         cfg.OAuth2.Mode,
	}
}

func (c *OAuth2Config) CallbackURI() string {
	return fmt.Sprintf("http://%s%s", c.ServeAddr, callbackURI)
}

func fixOAuthConfig(c *OAuth2Config) error {
	if c == nil {
		return errors.New("empty config")
	}

	if c.ServeAddr == "" {
		c.ServeAddr = "localhost:2333"
	}

	if c.AppID == "" || c.AppSecret == "" {
		return errNilOAuth2Application
	}

	if c.Mode != types.OAuth2Mode_Auto && c.Mode != types.OAuth2Mode_Manual {
		c.Mode = types.OAuth2Mode_Auto
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
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func NewOAuth2Support(c *OAuth2Config) IGitlabOauth2Support {
	if err := fixOAuthConfig(c); err != nil {
		panic(err)
	}

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

var (
	//go:embed callback.txt.tmpl
	callbackTxtTmpl embed.FS

	//go:embed callback.html.tmpl
	callbackHtmlTmpl embed.FS
)

func (g *gitlabOAuth2Support) renderCallback(w http.ResponseWriter, data interface{}) {
	var fs = callbackHtmlTmpl
	if g.oc.Mode == types.OAuth2Mode_Manual {
		fs = callbackTxtTmpl
	}

	var (
		tmpl interface {
			Execute(w io.Writer, data interface{}) error
		}
		err error
	)

	switch g.oc.Mode {
	case types.OAuth2Mode_Manual:
		tmpl, err = texttpl.ParseFS(fs, "callback.txt.tmpl")
	case types.OAuth2Mode_Auto:
		tmpl, err = htmltpl.ParseFS(fs, "callback.html.tmpl")
	default:
		_, _ = fmt.Fprintf(w, "invalid mode: %s", g.oc.Mode)
		return
	}

	if err != nil {
		log.Errorf("gitlabOAuth2Support.renderCallback parse FS(%s) failed: %v", g.oc.Mode, err)
		return
	}

	if err = tmpl.Execute(w, data); err != nil {
		log.Errorf("gitlabOAuth2Support.renderCallback failed to render: %v", err)
	}

	return
}

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

	// authorization callback is in line with the forecast.
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
	g.renderCallback(w, data)
}

// serve is serving a backend HTTP server process to
// receive redirect requests from gitlab.
func (g *gitlabOAuth2Support) serve() {
	http.HandleFunc(callbackURI, g.callbackHandl)
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
	_ = ctx
	form := url.Values{}

	form.Add("client_id", g.oc.AppID)
	form.Add("redirect_uri", g.oc.CallbackURI())
	form.Add("response_type", "code")
	form.Add("state", g.generateState())
	form.Add("scope", g.oc.Scopes)

	uri := fmt.Sprintf("%s%s?%s", g.oc.Host, step1URI, form.Encode())
	fmt.Printf("Your access token is invalid or expired, "+
		"please click following link to authorize: \n\t %s\n", uri)

	// oauth mode(OAuthMode_Manual) would not open a browser,
	// print the uri and tips here.
	if g.oc.Mode == types.OAuth2Mode_Manual {
		tips := "HINT: copy the link above and paste it into your browser to authorize.\n" +
			"Then copy the callback url from browser and execute as following command: \n\n" +
			"curl -X GET ${CALLBACK_URL}"
		fmt.Println(tips)
		return
	}

	// oauth mode(OAuthMode_Auto)
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
	form.Add("client_id", g.oc.AppID)
	form.Add("client_secret", g.oc.AppSecret)
	form.Add("redirect_uri", g.oc.CallbackURI())

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
	_ = ctx
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
