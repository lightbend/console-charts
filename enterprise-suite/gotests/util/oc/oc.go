package oc

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lightbend/gotests/util"
)

type Routes struct {
	ApiVersion string      `json:"apiVersion,omitempty"`
	Items      []Route     `json:"items,omitempty"`
	Kind       string      `json:"kind,omitempty"`
	Metadata   interface{} `json:"metadata,omitempty"`
}

type Route struct {
	ApiVersion string        `json:"apiVersion,omitempty"`
	Kind       string        `json:"kind,omitempty"`
	Metadata   RouteMetadata `json:"metadata,omitempty"`
	Spec       RouteSpec     `json:"spec,omitempty"`
	Status     RouteStatus   `json:"status,omitempty"`
}

type RouteMetadata struct {
	Annotations       map[string]string `json:"annotations,omitempty"`
	CreationTimestamp string            `json:"creationTimestamp,omitempty"`
	Name              string            `json:"name,omitempty"`
	Namespace         string            `json:"namespace,omitempty"`
	ResourceVersion   string            `json:"resourceVersion,omitempty"`
	SelfLink          string            `json:"selfLink,omitempty"`
	Uid               string            `json:"uid,omitempty"`
}

type RouteSpec struct {
	Host           string                 `json:"host,omitempty"`
	Port           map[string]string      `json:"port,omitempty"`
	To             map[string]interface{} `json:"to,omitempty"`
	WildcardPolicy string                 `json:"wildcardPolicy,omitempty"`
}

type RouteStatus struct {
	Ingress []interface{} `json:"ingress,omitempty"`
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

func Unexpose(service string) error {
	if _, err := util.Cmd("oc", "delete", "route", service).Run(); err != nil {
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

	for _, r := range routes.Items {
		if r.Metadata.Name == service {
			return r.Spec.Host, nil
		}
	}

	return "", fmt.Errorf("didn't find routes for service %v", service)
}
