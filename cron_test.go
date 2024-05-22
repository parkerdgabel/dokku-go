package dokku

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type cronManagerTestSuite struct {
	dokkuTestSuite
}

func TestRunCronManagerTestSuite(t *testing.T) {
	suite.Run(t, new(cronManagerTestSuite))
}

func (s *cronManagerTestSuite) TestCronOutputParse() {
	r := s.Suite.Require()

	output := `ID                                    Schedule   Command
cGhwPT09cGhwIHRlc3QucGhwPT09QGRhaWx5  @daily     node index.js
cGhwPT09dHJ1ZT09PSogKiAqICogKg==      * * * * *  true`
	task1 := CronTask{
		ID:       "cGhwPT09cGhwIHRlc3QucGhwPT09QGRhaWx5",
		Schedule: "@daily",
		Command:  "node index.js",
	}
	task2 := CronTask{
		ID:       "cGhwPT09dHJ1ZT09PSogKiAqICogKg==",
		Schedule: "* * * * *",
		Command:  "true",
	}
	expectedCrons := []CronTask{task1, task2}
	crons, err := parseCronOutput(output)
	r.NoError(err, "failed to parse cron output")
	r.EqualValues(expectedCrons, crons)
}
