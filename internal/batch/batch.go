package batch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Batch struct {
	batchUrls []string
}

func (b *Batch) SetBatchUrls(batchUrls []string) {
	b.batchUrls = batchUrls
}

func (b *Batch) request(path string, body io.Reader) ([]byte, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	for _, batchUrl := range b.batchUrls {
		url := batchUrl + path

		resp, err := client.Post(url, "application/json", body)
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

func (b *Batch) toTime(data []byte) (*time.Time, error) {
	var result map[string]interface{}

	var unmarshalErr = json.Unmarshal([]byte(string(data)), &result)

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	layout := "2006-01-02T15:04:05Z"
	t, parseErr := time.Parse(layout, result["data"].(string))

	if parseErr != nil {
		return nil, parseErr
	}

	return &t, nil
}

func (b *Batch) GetPreNextSchedule(name string, payload map[string]string) (*time.Time, error) {
	path := fmt.Sprintf("/api/v1/schedule/pre-next/%s", name)
	body, marshalErr := json.Marshal(payload)

	if marshalErr != nil {
		return nil, marshalErr
	}

	data, reqErr := b.request(path, bytes.NewBuffer(body))

	if reqErr != nil {
		return nil, reqErr
	}

	return b.toTime(data)
}

func (b *Batch) GetNextSchedule(id string) (*time.Time, error) {
	path := fmt.Sprintf("/api/v1/schedule/next/%s", id)

	data, reqErr := b.request(path, nil)

	if reqErr != nil {
		return nil, reqErr
	}

	return b.toTime(data)
}
