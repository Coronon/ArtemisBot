package artemis

func (d *ExerciseDetails) GetMostRecentScore() int {
	if len(d.StudentParticipations[0].Results) == 0 {
		return 0
	}

	var mostRecentID int
	var mostRecentScore int
	for _, result := range d.StudentParticipations[0].Results {
		if result.ID > mostRecentID {
			mostRecentID = result.ID
			mostRecentScore = int(result.Score)
		}
	}

	return mostRecentScore
}
