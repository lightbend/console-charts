package oc

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lightbend/gotests/util"
)

type Routes struct {
	apiVersion string
	items      []Route
	kind       string
	metadata   interface{}
}

type Route struct {
	apiVersion string
	kind       string
	metadata   RouteMetadata
	spec       RouteSpec
	status     RouteStatus
}

type RouteMetadata struct {
	annotations       map[string]string
	creationTimestamp string
	name              string
	namespace         string
	resourceVersion   string
	selfLink          string
	uid               string
}

type RouteSpec struct {
	host           string
	port           map[string]string
	to             map[string]string
	wildcardPolicy string
}

type RouteStatus struct {
	ingress []interface{}
}

func IsRunning() bool {
	retcode, err := util.Cmd("oc", "status").Run()
	return retcode == 0 && err == nil
}

func Expose(service string) error {
	if _, err := util.Cmd("oc", "expose", "service", service).Run(); err != nil {
		return err
	}
	return nil
}

func Address(service string) (string, error) {
	var stdout strings.Builder
	cmd := util.Cmd("oc", "get", "route", "-o", "json").CaptureStdout(&stdout)
	if _, err := cmd.Run(); err != nil {
		return "", err
	}

	var routes Routes
	if err := json.Unmarshal([]byte(stdout.String()), &routes); err != nil {
		return "", err
	}

	for _, r := range routes.items {
		if r.metadata.name == service {
			return r.spec.host, nil
		}
	}

	return "", fmt.Errorf("didn't find route for service %v", service)
}
