package multiTenancyPlugins

import (
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"
	"github.com/multiTenancyPlugins/apifilter"
	"github.com/multiTenancyPlugins/authentication"
	"github.com/multiTenancyPlugins/authorization"
	"github.com/multiTenancyPlugins/dataInit"
	"github.com/multiTenancyPlugins/flavors"
	"github.com/multiTenancyPlugins/keystone"
	"github.com/multiTenancyPlugins/naming"
	"github.com/multiTenancyPlugins/pluginAPI"
	"github.com/multiTenancyPlugins/quota"
	"github.com/multiTenancyPlugins/utils"
)

//Executor - Entry point to multi-tenancy plugins
type Executor struct{}

var startHandler pluginAPI.Handler

//Handle - Hook point from primary to plugins
func (*Executor) Handle(cluster cluster.Cluster, swarmHandler http.Handler) http.Handler {
	if os.Getenv("SWARM_MULTI_TENANT") == "false" {
		return swarmHandler
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug(r)
		var errInfo utils.ErrorInfo
		errInfo = startHandler(utils.ParseCommand(r), cluster, w, r, swarmHandler)
		if errInfo.Err != nil {
			log.Error(errInfo.Err)
			http.Error(w, errInfo.Err.Error(), errInfo.Status)
		}
	})
}

//Init - Initialize the Validation and Handling plugins
func (*Executor) Init() {
	if os.Getenv("SWARM_MULTI_TENANT") == "false" {
		log.Debug("SWARM_MULTI_TENANT is false")
		return
	}
	quotaPlugin := quota.NewQuota(nil)
	authorizationPlugin := authorization.NewAuthorization(quotaPlugin.Handle)
	nameScoping := namescoping.NewNameScoping(authorizationPlugin.Handle)
	mappingPlugin := dataInit.NewMapping(nameScoping.Handle)
	flavorsPlugin := flavors.NewPlugin(mappingPlugin.Handle)
	apiFilterPlugin := apifilter.NewPlugin(flavorsPlugin.Handle)
	authenticationPlugin := authentication.NewAuthentication(apiFilterPlugin.Handle)
	keystonePlugin := keystone.NewPlugin(authenticationPlugin.Handle)
	startHandler = keystonePlugin.Handle
}
