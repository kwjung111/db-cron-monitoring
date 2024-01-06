package queries

import (
	"jobs"
	"log"
	"strconv"
)

func ExampleQuery() jobs.Job {

	query := `  -- 10분 내 발송 시작 안 된 것 
	SELECT
	  '1' AS FIRSTCOL
	  , '2' AS SECONDCOL
	FROM
	 DUAL`

	messages := []string{"FIRSTCOL IS : ", "SECONDCOL IS : "}
	columns := []string{"FIRSTCOL", "SECONDCOL"}

	// 모든 Row 에 적용될 filter 정의. true 면 통과, false 면 걸러짐
	filter := func(cols map[string]string) bool {
		secondcol, err := strconv.Atoi(cols["SECONDCOL"])
		if err != nil {
			log.Fatal("convert Failed ", err)
		}
		if secondcol == 1 {
			return false
		}

		return true
	}

	job := jobs.MakeJob().
		SetName("Example").
		SetHead("this is Example query!").
		SetCron("0/5 * * * * *").
		SetMessages(messages).
		SetColumns(columns).
		SetQuery(query).
		SetFilter(filter).
		Build()

	return job
}
