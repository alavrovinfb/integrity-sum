package splunkclient

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/integrity-sum/pkg/notification"
	"github.com/sirupsen/logrus"
)

type event struct {
	Message string
}

type eventHolder struct {
	Time  float64 `json:"time"`
	Event event   `json:"event"`
}

type splunkClient struct {
	logger *logrus.Logger
	client *http.Client
	uri    string
	token  string
}

func New(logger *logrus.Logger, uri string, token string, insecureSkipVerify bool) notification.Notifier {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
	}
	client := &http.Client{Transport: transCfg}

	return &splunkClient{
		logger: logger,
		client: client,
		uri:    uri,
		token:  token,
	}
}

func (c *splunkClient) Send(msg notification.Message) error {

	eh := eventHolder{
		Time: float64(msg.Time.UnixNano()) / 1e9,
		Event: event{
			Message: msg.Message,
		},
	}

	data, err := json.Marshal(eh)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.uri, bytes.NewReader(data))
	if err != nil {
		return err
	}
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

	c.logger.WithField("body", string(respData)).WithField("Status", resp.StatusCode).Debug("Response received")

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("failed send message: invalid status code")
}
