package pluginAPI

import (
	"net/http"

	"github.com/docker/swarm/cluster"
	"github.com/multiTenancyPlugins/utils"
)

//This type define plugin entry function signature.
type Handler func(command utils.CommandEnum, cluster cluster.Cluster, w http.ResponseWriter, r *http.Request, swarmHandler http.Handler) utils.ErrorInfo

type PluginAPI interface {
	Handle(command utils.CommandEnum, cluster cluster.Cluster, w http.ResponseWriter, r *http.Request, swarmHandler http.Handler) utils.ErrorInfo
}
