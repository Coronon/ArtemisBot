package artemis

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/sync/singleflight"

	"github.com/coronon/artemisbot/internal/config"
	"github.com/coronon/artemisbot/internal/sockjs"
)

type ArtemisClient struct {
	sf singleflight.Group

	Username string
	password string

	jwt  *jwt.Token
	HTTP *resty.Client
	WS   *sockjs.SockJSClient

	// Absolute path to the working directory
	WorkDir string
}

// Create a new authenticated Artemis client
func NewArtemisClient(username, password, workdir string) (*ArtemisClient, error) {
	client := ArtemisClient{
		sf: singleflight.Group{},

		Username: username,
		password: password,

		WorkDir: workdir,
	}

	client.HTTP = resty.New().
		OnBeforeRequest(buildClientAuthMiddleware(&client))

	log.Debug("Trying to authenticate with Artemis...")
	err := client.Authenticate(&AuthenticateRequest{
		Username:     username,
		Password:     password,
		RemememberMe: true,
	})
	if err != nil {
		return nil, err
	}
	log.Debug("Successfully authenticated with Artemis")

	log.Debug("Creating websocket connection to Artemis...")
	wsHeaders := http.Header{}
	wsHeaders.Add("Origin", config.C.ArtemisHttpURL)
	wsHeaders.Add("Cookie", fmt.Sprintf("jwt=%s", client.jwt.Raw))
	wsClient, err := sockjs.NewSockJSClient(
		fmt.Sprintf("%s/0/a/websocket", config.C.ArtemisWsURL),
		wsHeaders,
	)
	if err != nil {
		return nil, err
	}
	client.WS = wsClient
	log.Debug("Successfully connected to Artemis websocket")

	return &client, nil
}

// Shutdown the Artemis client (mainly the websocket connection)
func (c *ArtemisClient) Close() {
	c.WS.Close()
}

// Check if the client is authenticated
func (c *ArtemisClient) IsAuthenticated() bool {
	if c.jwt == nil {
		return false
	}

	expiryTime, err := c.jwt.Claims.GetExpirationTime()
	if err != nil {
		return false
	}

	return expiryTime.After(time.Now().Add(30 * time.Second))
}

func buildClientAuthMiddleware(artemisClient *ArtemisClient) resty.RequestMiddleware {
	return func(restyClient *resty.Client, request *resty.Request) error {
		if request.URL == config.C.ArtemisHttpURL+"/public/authenticate" {
			return nil
		}

		if !artemisClient.IsAuthenticated() {
			err := artemisClient.Authenticate(&AuthenticateRequest{
				Username: artemisClient.Username,
				Password: artemisClient.password,
			})
			if err != nil {
				return err
			}
		}

		request.SetCookie(&http.Cookie{
			Name:  "jwt",
			Value: artemisClient.jwt.Raw,
		})

		return nil
	}
}

// Get the path to a new temporary directory and optionally create it
//
// The directory containing this path is guaranteed to exist.
func (c *ArtemisClient) NewTempDir(create bool) (string, error) {
	os.MkdirAll(c.WorkDir, 0755)

	dir, err := os.MkdirTemp(c.WorkDir, "artemisbot-")
	if err != nil {
		return "", err
	}

	if !create {
		os.Remove(dir)
	}

	return dir, nil
}
