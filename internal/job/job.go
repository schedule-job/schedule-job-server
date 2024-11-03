package job

import (
	"log"

	"github.com/schedule-job/schedule-job-database/core"
	"github.com/schedule-job/schedule-job-server/internal/errorset"
)

type Job struct {
	db core.Database
}

type Info struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Members     []string `json:"members"`
}

type Item struct {
	Name    string                 `json:"name"`
	Payload map[string]interface{} `json:"payload"`
}

type InsertItem struct {
	Info    Info `json:"info"`
	Action  Item `json:"action"`
	Trigger Item `json:"trigger"`
}

func (j *Job) SetDatabase(db core.Database) {
	j.db = db
}

func (j *Job) InsertJob(item InsertItem) (string, error) {
	id, err := j.db.InsertJob(item.Info.Name, item.Info.Description, item.Info.Author, item.Info.Members)

	if err != nil {
		log.Fatalln(err.Error())
		return "", errorset.ErrDatabase
	}

	errAction := j.db.InsertAction(id, item.Action.Name, item.Action.Payload)

	if errAction != nil {
		j.db.DeleteJob(id)
		log.Fatalln(errAction.Error())
		return "", errorset.ErrDatabase
	}

	errTrigger := j.db.InsertTrigger(id, item.Trigger.Name, item.Trigger.Payload)

	if errTrigger != nil {
		j.db.DeleteAction(id)
		j.db.DeleteJob(id)
		log.Fatalln(errTrigger.Error())
		return "", errorset.ErrDatabase
	}

	return id, nil
}

func (j *Job) DeleteJob(job_id string) {
	j.db.DeleteAction(job_id)
	j.db.DeleteTrigger(job_id)
	j.db.DeleteJob(job_id)
}
