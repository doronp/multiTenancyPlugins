package flavors

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"
	clusterParams "github.com/docker/swarm/cluster"
	"github.com/multiTenancyPlugins/pluginAPI"
	"github.com/multiTenancyPlugins/utils"
)

type DefaultFlavorsImpl struct {
	nextHandler pluginAPI.Handler
}

func NewPlugin(handler pluginAPI.Handler) pluginAPI.PluginAPI {
	flavorsPlugin := &DefaultFlavorsImpl{
		nextHandler: handler,
	}
	return flavorsPlugin
}

const MEGABYTE = 1048576

type Flavor struct {
	Memory int64
}

var flavors map[string]Flavor
var flavorsEnforced = os.Getenv("SWARM_FLAVORS_ENFORCED")

func init() {
	readFlavorFile()

}
func readFlavorFile() {
	if flavorsEnforced != "true" {
		log.Debug("Flavors not enforced")
		return
	}
	var flavorsFile = os.Getenv("SWARM_FLAVORS_FILE")
	if flavorsFile == "" {
		log.Debug("Missing SWARM_FLAVORS_FILE environment variable, using locate default ./flavors.json")
		flavorsFile = "flavors.json"
	}

	file, err := os.Open(flavorsFile)
	if err != nil {
		log.Fatal(err)
		panic("Error: could not open flavorsFile ")
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&flavors)
	if err != nil {
		log.Fatal("Error in flavors file decode:", err)
		panic("Error: could not decode flavors file ")
	}
	if _, ok := flavors["default"]; !ok {
		log.Fatal("Error flavors file does not contain default flavor")
		panic("Error: flavors file does not contain default flavor")
	}
	// convert memory to megabytes
	for key, value := range flavors {
		flavors[key] = Flavor{value.Memory * MEGABYTE}
	}
	log.Debugf("Flavors %+v", flavors)
}
func (flavorsImpl *DefaultFlavorsImpl) Handle(command utils.CommandEnum, cluster cluster.Cluster, w http.ResponseWriter, r *http.Request, swarmHandler http.Handler) utils.ErrorInfo {
	var errInfo utils.ErrorInfo
	errInfo.Status = http.StatusBadRequest
	if flavorsEnforced != "true" {
		return flavorsImpl.nextHandler(command, cluster, w, r, swarmHandler)
	}
	log.Debug("Plugin flavors Got command: " + command)
	if command != utils.CONTAINER_CREATE {
		return flavorsImpl.nextHandler(command, cluster, w, r, swarmHandler)
	}
	defer r.Body.Close()
	if reqBody, _ := ioutil.ReadAll(r.Body); len(reqBody) > 0 {
		var flavorIn Flavor
		var buf bytes.Buffer
		var oldconfig clusterParams.OldContainerConfig
		if err := json.NewDecoder(bytes.NewReader(reqBody)).Decode(&oldconfig); err != nil {
			errInfo.Err = err
			return errInfo
		}

		// make sure HostConfig fields are consolidated before creating container
		clusterParams.ConsolidateResourceFields(&oldconfig)

		flavorIn.Memory = oldconfig.ContainerConfig.HostConfig.Memory
		curKey := "default"
		for key, value := range flavors {
			if value == flavorIn {
				curKey = key
				break
			}
		}
		log.Debug("Plugin flavors apply flavor: ", curKey)

		oldconfig.ContainerConfig.HostConfig.Memory = flavors[curKey].Memory

		if err := json.NewEncoder(&buf).Encode(oldconfig); err != nil {
			errInfo.Err = err
			return errInfo
		}
		r, _ = utils.ModifyRequest(r, bytes.NewReader(buf.Bytes()), "", "")
		return flavorsImpl.nextHandler(command, cluster, w, r, swarmHandler)
	}
	errInfo.Err = errors.New("Plugin flavors enforced but returning nil!")
	return errInfo
}
