package main

import (
    "errors"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/entities"
)

var Priority = 1006  // must have higher prio than ratelimiting plugin to make sure it's respected

type Config struct {
	IntrospectionEndpoint string `json:"introspection_endpoint"`
	IntrospectionClientCredentials string `json:"introspection_client_credentials"`
}

type IntrospectionResult struct {
	Active bool `json:"active"`
	ClientId string `json:"client_id"`
}

func New() interface{} {
	return &Config{}
}

func (conf Config) Access(kong *pdk.PDK) {
	authHeader, err := kong.Request.GetHeader("authorization")
	if err != nil {
		kong.Log.Err(err.Error())
		return
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
	    kong.Log.Err("No bearer token received")
	    return
	}

    token := authHeader[len("Bearer "):]
	clientId, err := Introspect(conf.IntrospectionEndpoint, conf.IntrospectionClientCredentials, token)
	if len(clientId) > 0 {
        consumer := entities.Consumer{Id: clientId}
        kong.Client.Authenticate(&consumer, nil)
        kong.Log.Info(fmt.Sprintf("Set %s as request consumer", clientId))
        return
    }

    kong.Log.Err(err.Error())
    return
}

func Introspect(endpoint string, credentials string, token string) (string, error) {
    client := &http.Client{}

    req, err := http.NewRequest("POST", endpoint, strings.NewReader(fmt.Sprintf("token=%s", token)))
    req.Header.Add("Authorization", fmt.Sprintf("Basic %s", credentials))
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    resp, err := client.Do(req)

    if err != nil {
        return "", errors.New(fmt.Sprintf("introspection request failed: %s", err.Error()))
    }

    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        responseData, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            return "", errors.New("Failed to read introspection response")
        }
        return "", errors.New(fmt.Sprintf("introspection request, unexpected status: <%d>: %s", resp.StatusCode, string(responseData)))
    }

    var result IntrospectionResult
    err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
        return "", errors.New(fmt.Sprintf("introspection response could not be parsed: %s", err.Error()))
    }

    if !result.Active {
        return "", errors.New("token is not active")
    }

    return result.ClientId, nil
}