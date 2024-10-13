package job

import (
	"github.com/schedule-job/schedule-job-server/internal/pg"
)

type Job struct {
	db *pg.PostgresSQL
}

type info struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Members     []string `json:"members"`
}

type item struct {
	Name    string                 `json:"name"`
	Payload map[string]interface{} `json:"payload"`
}

type InsertItem struct {
	Info    info `json:"info"`
	Action  item `json:"action"`
	Trigger item `json:"trigger"`
}

func (j *Job) SetDatabase(db *pg.PostgresSQL) {
	j.db = db
}

func (j *Job) InsertJob(item InsertItem) (string, error) {
	id, err := j.db.InsertJob(item.Info.Name, item.Info.Description, item.Info.Author, item.Info.Members)

	if err != nil {
		return "", err
	}

	actionErr := j.db.InsertAction(id, item.Action.Name, item.Action.Payload)

	if actionErr != nil {
		j.db.DeleteJob(id)
		return "", actionErr
	}

	triggerErr := j.db.InsertTrigger(id, item.Trigger.Name, item.Trigger.Payload)

	if triggerErr != nil {
		j.db.DeleteAction(id)
		j.db.DeleteJob(id)
		return "", triggerErr
	}

	return id, nil
}
