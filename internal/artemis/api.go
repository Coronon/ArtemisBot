package artemis

import (
	"fmt"
	"time"

	"github.com/coronon/artemisbot/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type AuthenticateRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	RemememberMe bool   `json:"rememberMe"`
}

// Authenticate with Artemis and store the JWT in the client
func (c *ArtemisClient) Authenticate(req *AuthenticateRequest) error {
	_, err, _ := c.sf.Do("authenticate", func() (interface{}, error) {
		resp, err := c.HTTP.R().
			SetBody(req).
			Post(config.C.ArtemisHttpURL + "/public/authenticate")
		if err != nil {
			return nil, err
		}

		if resp.StatusCode() != 200 {
			return nil, fmt.Errorf("failed to authenticate with Artemis: %s", resp.Status())
		}

		var foundJWT string
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "jwt" {
				foundJWT = cookie.Value
				break
			}
		}

		if foundJWT == "" {
			return nil, fmt.Errorf("no JWT found in response")
		}

		token, _, err := jwt.NewParser().ParseUnverified(foundJWT, &jwt.RegisteredClaims{})
		if err != nil {
			return nil, err
		}
		token.Valid = true

		c.jwt = token

		return nil, nil
	})

	return err
}

type ExerciseDetails struct {
	Type                                   string    `json:"type"`
	ID                                     int       `json:"id"`
	Title                                  string    `json:"title"`
	ShortName                              string    `json:"shortName"`
	MaxPoints                              float64   `json:"maxPoints"`
	BonusPoints                            float64   `json:"bonusPoints"`
	AssessmentType                         string    `json:"assessmentType"`
	ReleaseDate                            time.Time `json:"releaseDate"`
	DueDate                                time.Time `json:"dueDate"`
	ExampleSolutionPublicationDate         time.Time `json:"exampleSolutionPublicationDate"`
	Difficulty                             string    `json:"difficulty"`
	Mode                                   string    `json:"mode"`
	AllowComplaintsForAutomaticAssessments bool      `json:"allowComplaintsForAutomaticAssessments"`
	AllowManualFeedbackRequests            bool      `json:"allowManualFeedbackRequests"`
	IncludedInOverallScore                 string    `json:"includedInOverallScore"`
	ProblemStatement                       string    `json:"problemStatement"`
	Categories                             []string  `json:"categories"`
	PresentationScoreEnabled               bool      `json:"presentationScoreEnabled"`
	SecondCorrectionEnabled                bool      `json:"secondCorrectionEnabled"`
	Course                                 struct {
		ID                                             int       `json:"id"`
		Title                                          string    `json:"title"`
		ShortName                                      string    `json:"shortName"`
		StudentGroupName                               string    `json:"studentGroupName"`
		TeachingAssistantGroupName                     string    `json:"teachingAssistantGroupName"`
		EditorGroupName                                string    `json:"editorGroupName"`
		InstructorGroupName                            string    `json:"instructorGroupName"`
		StartDate                                      time.Time `json:"startDate"`
		EndDate                                        time.Time `json:"endDate"`
		EnrollmentStartDate                            time.Time `json:"enrollmentStartDate"`
		EnrollmentEndDate                              time.Time `json:"enrollmentEndDate"`
		Semester                                       string    `json:"semester"`
		TestCourse                                     bool      `json:"testCourse"`
		DefaultProgrammingLanguage                     string    `json:"defaultProgrammingLanguage"`
		OnlineCourse                                   bool      `json:"onlineCourse"`
		CourseInformationSharingConfiguration          string    `json:"courseInformationSharingConfiguration"`
		CourseInformationSharingMessagingCodeOfConduct string    `json:"courseInformationSharingMessagingCodeOfConduct"`
		MaxComplaints                                  int       `json:"maxComplaints"`
		MaxTeamComplaints                              int       `json:"maxTeamComplaints"`
		MaxComplaintTimeDays                           int       `json:"maxComplaintTimeDays"`
		MaxRequestMoreFeedbackTimeDays                 int       `json:"maxRequestMoreFeedbackTimeDays"`
		MaxComplaintTextLimit                          int       `json:"maxComplaintTextLimit"`
		MaxComplaintResponseTextLimit                  int       `json:"maxComplaintResponseTextLimit"`
		EnrollmentEnabled                              bool      `json:"enrollmentEnabled"`
		UnenrollmentEnabled                            bool      `json:"unenrollmentEnabled"`
		AccuracyOfScores                               int       `json:"accuracyOfScores"`
		RestrictedAthenaModulesAccess                  bool      `json:"restrictedAthenaModulesAccess"`
		TimeZone                                       string    `json:"timeZone"`
		LearningPathsEnabled                           bool      `json:"learningPathsEnabled"`
		RequestMoreFeedbackEnabled                     bool      `json:"requestMoreFeedbackEnabled"`
		ComplaintsEnabled                              bool      `json:"complaintsEnabled"`
	} `json:"course"`
	StudentParticipations []struct {
		Type                string    `json:"type"`
		ID                  int       `json:"id"`
		InitializationState string    `json:"initializationState"`
		InitializationDate  time.Time `json:"initializationDate"`
		TestRun             bool      `json:"testRun"`
		Results             []struct {
			ID             int       `json:"id"`
			CompletionDate time.Time `json:"completionDate"`
			Successful     bool      `json:"successful"`
			Score          float64   `json:"score"`
			Rated          bool      `json:"rated"`
			Submission     struct {
				SubmissionExerciseType string    `json:"submissionExerciseType"`
				ID                     int       `json:"id"`
				Submitted              bool      `json:"submitted"`
				Type                   string    `json:"type"`
				SubmissionDate         time.Time `json:"submissionDate"`
				CommitHash             string    `json:"commitHash"`
				BuildFailed            bool      `json:"buildFailed"`
				BuildArtifact          bool      `json:"buildArtifact"`
				Empty                  bool      `json:"empty"`
				DurationInMinutes      int       `json:"durationInMinutes"`
			} `json:"submission"`
			AssessmentType      string `json:"assessmentType"`
			TestCaseCount       int    `json:"testCaseCount"`
			PassedTestCaseCount int    `json:"passedTestCaseCount"`
			CodeIssueCount      int    `json:"codeIssueCount"`
		} `json:"results"`
		Submissions []struct {
			SubmissionExerciseType string    `json:"submissionExerciseType"`
			ID                     int       `json:"id"`
			Submitted              bool      `json:"submitted"`
			Type                   string    `json:"type"`
			SubmissionDate         time.Time `json:"submissionDate"`
			CommitHash             string    `json:"commitHash"`
			BuildFailed            bool      `json:"buildFailed"`
			BuildArtifact          bool      `json:"buildArtifact"`
			Empty                  bool      `json:"empty"`
			DurationInMinutes      int       `json:"durationInMinutes"`
		} `json:"submissions"`
		Student struct {
			ID                    int       `json:"id"`
			CreatedDate           time.Time `json:"createdDate"`
			Login                 string    `json:"login"`
			FirstName             string    `json:"firstName"`
			LastName              string    `json:"lastName"`
			Email                 string    `json:"email"`
			Activated             bool      `json:"activated"`
			LangKey               string    `json:"langKey"`
			LastNotificationRead  time.Time `json:"lastNotificationRead"`
			Internal              bool      `json:"internal"`
			Name                  string    `json:"name"`
			ParticipantIdentifier string    `json:"participantIdentifier"`
			Deleted               bool      `json:"deleted"`
		} `json:"student"`
		RepositoryURI                string `json:"repositoryUri"`
		BuildPlanID                  string `json:"buildPlanId"`
		Branch                       string `json:"branch"`
		Locked                       bool   `json:"locked"`
		UserIndependentRepositoryURI string `json:"userIndependentRepositoryUri"`
		ParticipantIdentifier        string `json:"participantIdentifier"`
		ParticipantName              string `json:"participantName"`
	} `json:"studentParticipations"`
	AllowOnlineEditor               bool   `json:"allowOnlineEditor"`
	AllowOfflineIde                 bool   `json:"allowOfflineIde"`
	StaticCodeAnalysisEnabled       bool   `json:"staticCodeAnalysisEnabled"`
	ProgrammingLanguage             string `json:"programmingLanguage"`
	SequentialTestRuns              bool   `json:"sequentialTestRuns"`
	ShowTestNamesToStudents         bool   `json:"showTestNamesToStudents"`
	TestCasesChanged                bool   `json:"testCasesChanged"`
	ProjectKey                      string `json:"projectKey"`
	ProjectType                     string `json:"projectType"`
	TestwiseCoverageEnabled         bool   `json:"testwiseCoverageEnabled"`
	ReleaseTestsWithExampleSolution bool   `json:"releaseTestsWithExampleSolution"`
	CheckoutSolutionRepository      bool   `json:"checkoutSolutionRepository"`
	ExerciseType                    string `json:"exerciseType"`
	StudentAssignedTeamIDComputed   bool   `json:"studentAssignedTeamIdComputed"`
	GradingInstructionFeedbackUsed  bool   `json:"gradingInstructionFeedbackUsed"`
	TeamMode                        bool   `json:"teamMode"`
	VisibleToStudents               bool   `json:"visibleToStudents"`
}

// Get the details of an exercise on Artemis by its ID
func (c *ArtemisClient) GetExerciseDetails(exerciseID string) (*ExerciseDetails, error) {
	details, err, _ := c.sf.Do(fmt.Sprintf("exercise-details-%s", exerciseID), func() (interface{}, error) {
		var details ExerciseDetails
		resp, err := c.HTTP.R().
			SetResult(&details).
			Get(fmt.Sprintf(
				"%s/exercises/%s/details",
				config.C.ArtemisHttpURL,
				exerciseID,
			))
		if err != nil {
			return nil, err
		}

		if resp.StatusCode() != 200 {
			return nil, fmt.Errorf("failed to get exercise details: %s", resp.Status())
		}

		return &details, nil
	})

	return details.(*ExerciseDetails), err
}
