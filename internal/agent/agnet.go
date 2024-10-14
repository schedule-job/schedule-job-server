package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/schedule-job/schedule-job-server/internal/errorset"
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

		resp, errResp := client.Get(url)
		if errResp != nil {
			if errResp == context.DeadlineExceeded {
				continue
			}
			log.Fatalln(errResp.Error())
			return nil, errResp
		}

		defer resp.Body.Close()

		body, errRead := io.ReadAll(resp.Body)
		if errRead != nil {
			log.Fatalln(errRead.Error())
			return nil, errRead
		}

		return body, nil
	}

	log.Fatalln("no agent url")
	return nil, errorset.ErrInternalServer
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
		log.Fatalln(err.Error())
		return nil, errorset.ErrInternalServer
	}

	return resp, nil
}

func (a *Agent) GetLog(jobId, id string) ([]byte, error) {
	path := fmt.Sprintf("/api/v1/request/%s/log/%s", jobId, id)

	resp, err := a.request(path)

	if err != nil {
		log.Fatalln(err.Error())
		return nil, errorset.ErrInternalServer
	}

	return resp, nil
}
