package main

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/jhillyerd/enmime"
	"github.com/mhale/smtpd"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Port      string            `yaml:"port"`
	Senders   map[string]bool   `yaml:"senders"`
	Receivers map[string]string `yaml:"receivers"`
}

var config Config
var configFile string

func LoadConfig(configFile string) error {
	file, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, 1024)

	count, err := file.Read(buffer)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(buffer[:count], &config)
	if err != nil {
		return err
	}
	return nil
}

func sendFlockAlert(url, message string) {
	postBody, _ := json.Marshal(map[string]string{
		"text": message,
	})
	requestBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(url, "application/json", requestBody)
	if err != nil {
		log.Fatal().Msgf("Failed to send flock alert: %s", err)
	}
	defer resp.Body.Close()
	log.Info().Msgf("Notification Sent to flock, response status %d", resp.StatusCode)
}

func mailHandler(origin net.Addr, from string, to []string, data []byte) error {

	// 0. Read Text from mail body
	env, err := enmime.ReadEnvelope(bytes.NewReader(data))
	if err != nil {
		log.Error().Err(err)
	}

	// 1. Check if mail is sent from validation_senders
	if v, ok := config.Senders[from]; !(ok && v) {
		log.Info().Msgf("Sender validation failed received mail from %s for %s. IP: %s", from, to[0], origin.String())
		return nil
	}

	// 2. Check if receivers are in validation_receivers
	url, ok := config.Receivers[to[0]]
	if !ok {
		log.Info().Msgf("Receiver validation failed received mail from %s for %s. IP: %s", from, to[0], origin.String())
		return nil
	}

	// 3. Check if Text is not empty
	if len(env.Text) == 0 {
		log.Info().Msgf("Text validation failed received mail from %s for %s. IP: %s", from, to[0], origin.String())
		return nil
	}

	sendFlockAlert(url, env.Text)
	log.Info().Msgf("Processed mail from %s for %s. IP: %s", from, to[0], origin.String())

	return nil
}

func main() {
	// Loadconfig file
	configFile = "config.yaml"
	err := LoadConfig(configFile)
	if err != nil {
		log.Fatal().Msgf("Failed to read config: %s", err)
	}
	log.Info().Msgf("Starting Server in %s", config.Port)
	smtpd.ListenAndServe(config.Port, mailHandler, "SMTPProxyApp", "")
}
