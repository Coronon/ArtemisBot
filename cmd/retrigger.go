package cmd

import (
	"regexp"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/coronon/artemisbot/internal/artemis"
	"github.com/coronon/artemisbot/internal/git"
	"github.com/coronon/artemisbot/internal/sockjs"
	"github.com/coronon/artemisbot/internal/util"
)

var artemisURLRegex = regexp.MustCompile(`^https?://.+/courses/(?P<course>\d*)/exercises/(?P<task>\d*)/?`)
var expectingResult = true
var isExiting = false

// retriggerCmd represents the retrigger command
var retriggerCmd = &cobra.Command{
	Use:   "retrigger",
	Short: "Retrigger artemis build tasks",
	Long:  `Trigger the artemis build task until the desired percentage is reached.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		log.Info("Starting ArtemisBot 🤖")

		// Get options
		username := viper.GetString("username")
		password := viper.GetString("password")
		desiredPercentage := viper.GetInt("percentage")
		artemisURL := viper.GetString("artemis-url")
		isInteractive := viper.GetBool("interactive")
		workDir := viper.GetString("workdir")
		verbose := viper.GetBool("verbose")

		// Set log level
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
		log.SetLevel(log.DebugLevel)

		// Ensure that the username and password are set
		if isInteractive {
			log.Info("Please enter your Artemis credentials")
			username, password, err = util.GetCredentialsInteractive()
			if err != nil {
				log.Errorf("Could not read credentials: %s", err.Error())
				return
			}
		} else if username == "" || password == "" {
			log.Error("Username and password are required")
			return
		}

		// Extract courseID and exerciseID from the URL
		if artemisURL == "" {
			log.Error("Artemis URL is required")
			return
		}

		matches := artemisURLRegex.FindStringSubmatch(artemisURL)
		if len(matches) != 3 {
			log.Error("Failed to extract courseID and exerciseID from the URL")
			return
		}
		courseID := matches[1]
		taskID := matches[2]

		// Check if the desired percentage is valid
		if desiredPercentage < 0 {
			log.Error("The desired percentage must be greater than or equal to 0")
			return
		}
		if desiredPercentage == 0 {
			log.Info("The desired percentage is 0% 🤷‍♂️")
			return
		}

		// Start the loop
		shouldRunAgain := true
		for shouldRunAgain {
			shouldRunAgain = loop(username, password, workDir, courseID, taskID, desiredPercentage)

			if shouldRunAgain {
				log.Warn("Something went wrong, retrying in 5 seconds...")
				time.Sleep(5 * time.Second)
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(retriggerCmd)

	retriggerCmd.PersistentFlags().StringP("artemis-url", "t", "https://artemis.in.tum.de/courses/?/exercises/?", "URL of the Artemis task to automate")
	retriggerCmd.PersistentFlags().IntP("percentage", "p", 100, "Percentage of points to reach")
}

func loop(username, password, workDir, courseID, taskID string, desiredPercentage int) bool {
	log.Debug("Bootsrapping Artemis client...")
	// Create an Artemis client
	client, err := artemis.NewArtemisClient(username, password, workDir)
	if err != nil {
		log.Errorf("Could not create an Artemis client: %s", err.Error())
		return false
	}
	log.Debug("Artemis client bootstrapped")

	log.Debug("Creating a new Artemis task...")
	// Create a new Artemis task
	task, err := artemis.NewRetriggerTask(
		client,
		courseID,
		taskID,
		&git.GitCredentials{
			Username: username,
			Password: password,
		},
		desiredPercentage,
	)
	if err != nil {
		log.Errorf("Could not create a new Artemis task: %s", err.Error())
		return false
	}
	log.Debug("Artemis task created")

	// Ensure that the desired percentage is not already reached
	if task.CurrentPercentage >= task.DesiredPercentage {
		log.Info("The desired percentage is already reached 😎")
		return false
	}

	// Start listening for build events
	if err = client.WS.Subscribe("/user/topic/newSubmissions"); err != nil {
		log.Errorf("Could not subscribe to new submissions: %s", err.Error())
		return true
	}
	if err = client.WS.Subscribe("/user/topic/newResults"); err != nil {
		log.Errorf("Could not subscribe to new results: %s", err.Error())
		return true
	}

	log.Info("Starting the Artemis task... 🚀")
	timer := time.NewTimer(1 * time.Millisecond)
	for {
		select {
		case <-timer.C:
			// Retrigger the task
			err = retriggerTask(task)
			if err != nil {
				log.Errorf("Could not retrigger the task: %s", err.Error())
				return true
			}
		case msg := <-client.WS.Messages():
			err, shouldRetrigger := handleWSMessage(client, task, msg)
			if err != nil {
				log.Errorf("Could not handle websocket message: %s", err.Error())
				return true
			}

			if shouldRetrigger {
				timer.Reset(1 * time.Second)
			}
		case err := <-client.WS.Errors():
			log.Errorf("Websocket error: %s", err.Error())
		case <-client.WS.Done():
			log.Info("Websocket connection closed")
			return !isExiting
		}
	}
}

func retriggerTask(task *artemis.Task) error {
	log.Info("Retriggering the task... ⚙️")
	hash, err := task.Retrigger()
	if err != nil {
		return err
	}

	log.Infof("Retriggered the task with commit hash %s", hash)

	return nil
}

func handleWSMessage(client *artemis.ArtemisClient, task *artemis.Task, msg *sockjs.SockJSMessage) (error, bool) {
	if msg.Command != "MESSAGE" {
		return nil, false
	}

	switch msg.Headers["destination"] {
	case "/user/topic/newSubmissions":
		if expectingResult {
			log.Warn("Artemis had a little hiccup and sent a new submission before the results -> retrigger 🤔")
			return nil, true
		}

		log.Info("Artemis is building a new submission 📦")
		expectingResult = true
	case "/user/topic/newResults":
		newPercentage := int((*msg.Body)["score"].(float64))
		if newPercentage >= task.DesiredPercentage {
			log.Infof("The desired percentage is reached: %d%% 🎉", newPercentage)
			isExiting = true
			client.Close()
			return nil, false
		} else {
			log.Infof("Received new results: %d%%", newPercentage)
		}

		// Retrigger the task
		expectingResult = false
		return nil, true
	default:
		log.Debugf("Received message with unknown destination: %s", msg.Headers["destination"])
	}

	return nil, false
}
