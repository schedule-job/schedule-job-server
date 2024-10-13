package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Agent struct {
	agentUrls []string
}

func (a *Agent) SetAgentUrls(agentUrls []string) {
	a.agentUrls = agentUrls
}

func (a *Agent) request(path string) ([]byte, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	for _, agentUrl := range a.agentUrls {
		url := agentUrl + path

		resp, err := client.Get(url)
		if err != nil {
			if err == context.DeadlineExceeded {
				continue
			}
			return nil, err
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return body, nil
	}

	return nil, errors.New("no agent url")
}

func (a *Agent) GetLogs(jobId, lastId string, limit int) ([]byte, error) {
	query := []string{}
	if lastId != "" {
		query = append(query, "lastId="+lastId)
	}
	if limit > 0 {
		query = append(query, "limit="+strconv.Itoa(limit))
	}

	path := fmt.Sprintf("/api/v1/request/%s/logs?%s", jobId, strings.Join(query, "&"))

	resp, err := a.request(path)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (a *Agent) GetLog(jobId, id string) ([]byte, error) {
	path := fmt.Sprintf("/api/v1/request/%s/log/%s", jobId, id)

	resp, err := a.request(path)

	if err != nil {
		return nil, err
	}

	return resp, nil
}
