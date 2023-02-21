package splunk

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/ScienceSoft-Inc/integrity-sum/pkg/alerts"
)

type event struct {
	Message string `json:"message"`
	Reason  string `json:"reason"`
	Path    string `json:"path"`
}

type eventHolder struct {
	Time  float64 `json:"time"`
	Event event   `json:"event"`
}

type splunkClient struct {
	logger *logrus.Logger
	client *http.Client
	uri    string
	auth   string
}

func New(logger *logrus.Logger, uri string, token string, insecureSkipVerify bool) alerts.Sender {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
	}
	client := &http.Client{Transport: transCfg}

	return &splunkClient{
		logger: logger,
		client: client,
		uri:    uri,
		auth:   fmt.Sprintf("Splunk %v", token),
	}
}

func (c *splunkClient) Send(alert alerts.Alert) error {

	eh := eventFromAlert(alert)

	data, err := json.Marshal(eh)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.uri, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", c.auth)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			c.logger.WithError(err).Error("Failed close response body")
		}
	}()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	c.logger.WithField("body", string(respData)).WithField("Status", resp.StatusCode).Error("Failed send alert to splunk")

	return fmt.Errorf("failed send message: invalid status code")
}

func eventFromAlert(alert alerts.Alert) eventHolder {
	return eventHolder{
		Time: float64(alert.Time.UnixNano()) / 1e9,
		Event: event{
			Message: alert.Message,
			Reason:  alert.Reason,
			Path:    alert.Path,
		},
	}
}
