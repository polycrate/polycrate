package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type MonkConfig struct {
	Endpoint string `yaml:"endpoint,omitempty" mapstructure:"endpoint,omitempty" json:"endpoint,omitempty"`
	Enabled  bool   `yaml:"enabled,omitempty" mapstructure:"enabled,omitempty" json:"enabled,omitempty"`
}

type MonkEvent struct {
	Labels map[string]string `yaml:"Labels,omitempty" mapstructure:"Labels,omitempty" json:"Labels,omitempty"`
	Raw    []byte            `yaml:"Raw,omitempty" mapstructure:"Raw,omitempty" json:"Raw,omitempty"`
}

type Monk struct {
	Config MonkConfig `yaml:"config,omitempty" mapstructure:"config,omitempty" json:"config,omitempty"`
}

func (m *Monk) Load(ctx context.Context, config *MonkConfig) error {
	m.Config = *config
	return nil
}

func (m *Monk) SubmitEvent(ctx context.Context, data []byte) error {

	log.Warnf("Submitting event to Monk at %s", m.Config.Endpoint)

	event := MonkEvent{}
	event.Raw = data

	request, err := http.NewRequest("POST", m.Config.Endpoint, bytes.NewBuffer(event.Raw))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("response Body:", string(body))
	return nil

}

func (e *MonkEvent) AddLabel(ctx context.Context, key string, value string) error {
	e.Labels[key] = value
	return nil
}
