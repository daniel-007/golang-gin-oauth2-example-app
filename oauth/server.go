package oauth

import (
	"log"
	"net/http"

	"github.com/spf13/viper"

	"gopkg.in/go-oauth2/redis.v3"
	"gopkg.in/oauth2.v3"

	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"

	"sync"
	"time"

	"github.com/aristat/golang-gin-oauth2-example-app/common"
)

type oauth2Server struct {
	server *server.Server
}

var (
	IOauthServer common.OauthServer
	once         sync.Once
)

func GetIOauthServer() common.OauthServer {
	once.Do(func() {
		IOauthServer = NewOauthServer()
	})
	return IOauthServer
}

func NewOauthServer() common.OauthServer {
	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(
		&manage.Config{
			AccessTokenExp:    time.Hour * 24 * 7,
			RefreshTokenExp:   time.Hour * 24 * 14,
			IsGenerateRefresh: true,
		},
	)
	manager.MapTokenStorage(redis.NewRedisStore(&redis.Options{
		Addr: viper.GetString("REDIS_URL"),
		DB:   viper.GetInt("REDIS_TOKEN_DB"),
	}))

	clientStore := store.NewClientStore()
	for key, value := range clientsConfig {
		clientStore.Set(key, value)
	}

	manager.MapClientStorage(clientStore)

	oauthServer := &oauth2Server{server: server.NewDefaultServer(manager)}

	oauthServer.server.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	oauthServer.server.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	return oauthServer
}

func (m *oauth2Server) UserAuthorizationHandler(handler server.UserAuthorizationHandler) {
	m.server.UserAuthorizationHandler = handler
}

func (m *oauth2Server) HandleAuthorizeRequest(w http.ResponseWriter, r *http.Request) (err error) {
	err = m.server.HandleAuthorizeRequest(w, r)
	return
}

func (m *oauth2Server) HandleTokenRequest(w http.ResponseWriter, r *http.Request) (err error) {
	err = m.server.HandleTokenRequest(w, r)
	return
}

func (m *oauth2Server) ValidationBearerToken(r *http.Request) (ti oauth2.TokenInfo, err error) {
	ti, err = m.server.ValidationBearerToken(r)
	return
}
