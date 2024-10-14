package batch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/schedule-job/schedule-job-server/internal/errorset"
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

		resp, errResp := client.Post(url, "application/json", body)
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

	log.Fatalln("no batch url")
	return nil, errorset.ErrInternalServer
}

func (b *Batch) toTime(data []byte) (*time.Time, error) {
	var result map[string]interface{}

	var errUnmarshal = json.Unmarshal([]byte(string(data)), &result)

	if errUnmarshal != nil {
		log.Fatalln(errUnmarshal.Error())
		return nil, errUnmarshal
	}

	layout := "2006-01-02T15:04:05Z"
	t, errParse := time.Parse(layout, result["data"].(string))

	if errParse != nil {
		log.Fatalln(errParse.Error())
		return nil, errParse
	}

	return &t, nil
}

func (b *Batch) toJson(data []byte) (interface{}, error) {
	var result map[string]interface{}

	var errUnmarshal = json.Unmarshal([]byte(string(data)), &result)

	if errUnmarshal != nil {
		log.Fatalln(errUnmarshal.Error())
		return nil, errUnmarshal
	}

	return result["data"], nil
}

func (b *Batch) GetPreNextSchedule(name string, payload map[string]interface{}) (*time.Time, error) {
	path := fmt.Sprintf("/api/v1/schedule/pre-next/%s", name)
	body, errMarshal := json.Marshal(payload)

	if errMarshal != nil {
		log.Fatalln(errMarshal.Error())
		return nil, errMarshal
	}

	data, errReq := b.request(path, bytes.NewBuffer(body))

	if errReq != nil {
		log.Fatalln(errReq.Error())
		return nil, errorset.ErrInternalServer
	}

	time, errTime := b.toTime(data)

	if errTime != nil {
		log.Fatalln(errTime.Error())
		return nil, errorset.ErrInternalServer
	}

	return time, nil
}

func (b *Batch) GetNextSchedule(id string) (*time.Time, error) {
	path := fmt.Sprintf("/api/v1/schedule/next/%s", id)

	data, errReq := b.request(path, nil)

	if errReq != nil {
		log.Fatalln(errReq.Error())
		return nil, errorset.ErrInternalServer
	}

	time, errTime := b.toTime(data)

	if errTime != nil {
		log.Fatalln(errTime.Error())
		return nil, errorset.ErrInternalServer
	}

	return time, nil
}

func (b *Batch) GetPreNextInfo(name string, payload map[string]interface{}) (interface{}, error) {
	path := fmt.Sprintf("/api/v1/request/pre-next/%s", name)
	body, errMarshal := json.Marshal(payload)

	if errMarshal != nil {
		log.Fatalln(errMarshal.Error())
		return nil, errMarshal
	}

	data, errReq := b.request(path, bytes.NewBuffer(body))

	if errReq != nil {
		log.Fatalln(errReq.Error())
		return nil, errorset.ErrInternalServer
	}

	json, errJson := b.toJson(data)

	if errJson != nil {
		log.Fatalln(errJson.Error())
		return nil, errorset.ErrInternalServer
	}

	return json, nil
}

func (b *Batch) GetNextInfo(id string) (interface{}, error) {
	path := fmt.Sprintf("/api/v1/request/next/%s", id)

	data, errReq := b.request(path, nil)

	if errReq != nil {
		log.Fatalln(errReq.Error())
		return nil, errorset.ErrInternalServer
	}

	json, errJson := b.toJson(data)

	if errJson != nil {
		log.Fatalln(errJson.Error())
		return nil, errorset.ErrInternalServer
	}

	return json, nil
}

func (b *Batch) Progress() error {
	path := "/api/v1/progress"

	_, errReq := b.request(path, nil)

	if errReq != nil {
		log.Fatalln(errReq.Error())
		return errorset.ErrInternalServer
	}

	return nil
}

func (b *Batch) ProgressOnce(id string) error {
	path := fmt.Sprintf("/api/v1/progress/%s", id)

	_, errReq := b.request(path, nil)

	if errReq != nil {
		log.Fatalln(errReq.Error())
		return errorset.ErrInternalServer
	}

	return nil
}
