package artemis

import (
	"github.com/coronon/artemisbot/internal/easygit"
	"github.com/coronon/artemisbot/internal/git"
)

type Task struct {
	CourseID          string
	TaskID            string
	CurrentPercentage int
	DesiredPercentage int
	GitConfig         *git.GitConfig

	client         *ArtemisClient
	repository     git.Repository
	gitCredentials *git.GitCredentials
}

func NewRetriggerTask(
	client *ArtemisClient,
	courseID, taskID string,
	gitCredentials *git.GitCredentials,
	desiredPercentage int,
) (*Task, error) {
	task := &Task{
		CourseID:          courseID,
		TaskID:            taskID,
		DesiredPercentage: desiredPercentage,

		client:         client,
		gitCredentials: gitCredentials,
	}

	// Resolve the task
	if err := task.Resolve(); err != nil {
		return nil, err
	}

	// Clone the repository
	dir, err := task.client.NewTempDir(false)
	if err != nil {
		return nil, err
	}

	repo, err := easygit.NewRepository(task.GitConfig, task.gitCredentials, dir)
	if err != nil {
		return nil, err
	}
	task.repository = repo

	return task, nil
}

// Populate the task with the necessary information for working with the Artemis API
//
// Normally this should have already been done by the constructor.
func (t *Task) Resolve() error {
	details, err := t.client.GetExerciseDetails(t.TaskID)
	if err != nil {
		return err
	}

	// Git config
	t.GitConfig = &git.GitConfig{
		URL:    details.StudentParticipations[0].RepositoryURI,
		Branch: details.StudentParticipations[0].Branch,
		Name:   details.StudentParticipations[0].ParticipantName,
		Email:  details.StudentParticipations[0].ParticipantIdentifier + "@mytum.de",
	}

	// Current percentage
	t.CurrentPercentage = details.GetMostRecentScore()

	return nil
}

// Retrigger the task to update the percentage and return the commit hash
func (t *Task) Retrigger() (string, error) {
	return t.repository.PushEmptyCommit()
}
