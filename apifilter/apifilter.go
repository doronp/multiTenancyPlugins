package apifilter

import (
	"errors"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"
	"github.com/multiTenancyPlugins/pluginAPI"
	"github.com/multiTenancyPlugins/utils"
)

type DefaultApiFilterImpl struct {
	nextHandler pluginAPI.Handler
}

func NewPlugin(handler pluginAPI.Handler) pluginAPI.PluginAPI {
	apiFilterPlugin := &DefaultApiFilterImpl{
		nextHandler: handler,
	}
	return apiFilterPlugin
}

type Apifilter struct{}

func (apiFilterImpl *DefaultApiFilterImpl) Handle(command utils.CommandEnum, cluster cluster.Cluster, w http.ResponseWriter, r *http.Request, swarmHandler http.Handler) utils.ErrorInfo {
	log.Debug("Plugin apiFilter Got command: " + command)
	var errInfo utils.ErrorInfo
	errInfo.Status = http.StatusBadRequest
	if supportedAPIsMap[command] {
		return apiFilterImpl.nextHandler(command, cluster, w, r, swarmHandler)
	} else {
		errInfo.Err = errors.New("Command Not Supported!")
		return errInfo
	}

}

func init() {
	initSupportedAPIsMap()
	modifySupportedWithDisabledApi()

}
