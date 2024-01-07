package queries

import (
	"jobs"
	"strconv"
)

func ExampleQuery() jobs.Job {

	query := `  
	SELECT
	  '1' AS FIRSTCOL
	  ,'2' AS SECONDCOL
	FROM
	 DUAL`

	messages := []string{"FIRSTCOL IS : ", "SECONDCOL IS : "}
	columns := []string{"FIRSTCOL", "SECONDCOL"}

	// 모든 Row 에 적용될 filter => 모든 row 가 조건을 만족하면
	// 메세지를 보내지 않음 (true 면 통과, false 면 걸러짐)
	filter := func(cols map[string]string) (bool, error) {
		secondcol, err := strconv.Atoi(cols["SECONDCOL"])
		if err != nil {
			return false, err
		}
		if secondcol == 1 {
			return false, nil
		}

		return true, nil
	}

	textReplacer := func(cols map[string]string) (map[string]string, error) {

		newRow := cols
		newRow["SECONDCOL"] = "replaced"

		return newRow, nil
	}

	job := jobs.MakeJob().
		SetName("Example").
		SetHead("this is Example query!").
		SetCron("0/5 * * * * 0").
		SetMessages(messages).
		SetColumns(columns).
		SetQuery(query).
		SetFilter(filter).
		SetTextReplacer(textReplacer).
		Build()

	return job
}
