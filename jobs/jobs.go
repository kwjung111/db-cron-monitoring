package jobs

import (
	"connectionPool"
	"log"
	"sendMessage"
	"time"
	"util"

	"github.com/go-co-op/gocron/v2"
)

var defaultToken string

var jobLists = []*job{}

type job struct {
	name          string
	description   string
	query         string
	messages      []string
	columns       []string
	cron          string
	head          string
	filter        jobFilter
	textReplacer  jobTextReplacer
	recentStatus  string
	recentSuccess string
	token         string
}

type Job interface {
	Register(s gocron.Scheduler)
	Execute()
	makeBody([]map[string]string) (string, bool, error)
	GetDescription() string
	GetName() string
	GetCron() string
	GetRecentStatus() string
	GetRecentSuccess() string
	SetRecentStatus(string)
	SetRecentSuccess(string)
}

type jobBuilder struct {
	job job
}

type JobBuilder interface {
	SetName(string) JobBuilder
	SetQuery(string) JobBuilder
	SetHead(string) JobBuilder
	SetColumns([]string) JobBuilder
	SetMessages([]string) JobBuilder
	SetCron(string) JobBuilder
	SetFilter(jobFilter) JobBuilder
	SetTextReplacer(jobTextReplacer) JobBuilder
	SetDescription(string) JobBuilder
	Build() *job
}

type jobFilter func(map[string]string) (bool, error)

type jobTextReplacer func(map[string]string) (map[string]string, error)

// returns builder
func MakeJob() JobBuilder {
	return &jobBuilder{}
}

// Setter (required) : job.query
// 모니터링 쿼리 설정
func (jb *jobBuilder) SetQuery(query string) JobBuilder {
	jb.job.query = query
	return jb
}

// Setter (required) : job.name
// 에러 검출 및 로깅에 사용될 job 이름
func (jb *jobBuilder) SetName(name string) JobBuilder {
	jb.job.name = name
	return jb
}

// setter (optional) : job.head
// body 앞에 오는 설명
func (jb *jobBuilder) SetHead(head string) JobBuilder {
	jb.job.head = head + "\n"
	return jb
}

// setter (required) : job.columns
// message 사이에 끼워넣을 SQL 의 column 이름
// message[i] + columns[i] 의 순서로 message body 가 만들어짐
func (jb *jobBuilder) SetColumns(columns []string) JobBuilder {
	jb.job.columns = columns
	return jb
}

// setter (required) : job.messages
// column 값과 함께 message body를 구성하는 string 배열
func (jb *jobBuilder) SetMessages(messages []string) JobBuilder {
	jb.job.messages = messages
	return jb
}

// setter (required) : job.cron
// cron 표현식을 string type 으로 받음
func (jb *jobBuilder) SetCron(cron string) JobBuilder {
	jb.job.cron = cron
	return jb
}

// setter (optional) : filter
// map[string]string 타입의  에 적용될 filter 함수 설정
// 전체 row 에 대해 조건 만족 시 메세지가 전송되지 않음
func (jb *jobBuilder) SetFilter(filter jobFilter) JobBuilder {
	jb.job.filter = filter
	return jb
}

func (jb *jobBuilder) SetTextReplacer(replacer jobTextReplacer) JobBuilder {
	jb.job.textReplacer = replacer
	return jb
}

// setter (optional) : description
// 해당 job 에 대한 간략한 설명
func (jb *jobBuilder) SetDescription(description string) JobBuilder {
	jb.job.description = description
	return jb
}

// returns builded job
func (jb *jobBuilder) Build() *job {
	//validation -> name,cron
	if jb.job.name == "" {
		log.Println("job.name not defined")
	}
	jb.job.SetRecentStatus("PENDING")
	jb.job.SetRecentSuccess("no record")
	return &jb.job
}

// register job to main cron
func (j *job) Register(scheduler gocron.Scheduler) {
	scheduler.NewJob(
		gocron.CronJob(j.cron, true),
		gocron.NewTask(j.Execute),
	)
	PushJob(j)
}

// execute job
func (j *job) Execute() {

	rowMapList, err := util.GetResultfromDB(connectionPool.GetConnection(), j.query)
	if err != nil {
		log.Println("error : ", j.name, err)
		j.SetRecentStatus("FAILED")
		return
	}

	//메시지 생성 및 필터링
	head := j.head
	body, pass, err := j.makeBody(rowMapList)
	if err != nil {
		log.Println(j.name, " : ", err)
		j.SetRecentStatus("FAILED")
		return
	}
	if pass {
		token := j.token

		//Line API Token
		if token == "" {
			token = defaultToken
		}

		err = sendMessage.SendLineMessage(head+body, token)
		if err != nil {
			log.Println("error while sending Msg : ", err)
		}
	}

	now := time.Now()
	formattedTime := now.Format("2006-01-02 15:04:05")

	j.SetRecentStatus("SUCCEED")
	j.SetRecentSuccess(formattedTime)
}

// returns job.description
func (j *job) GetDescription() string {
	if j.description == "" {
		return j.name + " : have no description"
	}
	return j.description
}

// returns job.name
func (j *job) GetName() string {
	return j.name
}

// returns job.cron
func (j *job) GetCron() string {
	return j.cron
}

// returns job.recentStauts
func (j *job) GetRecentStatus() string {
	return j.recentStatus
}

// returns job.recentSucces
func (j *job) GetRecentSuccess() string {
	return j.recentSuccess
}

// set job.RenetStatus
func (j *job) SetRecentStatus(status string) {
	j.recentStatus = status
}

// set job.RenetStatus
func (j *job) SetRecentSuccess(time string) {
	j.recentSuccess = time
}

// push job to job list
func PushJob(j *job) {
	jobLists = append(jobLists, j)
}

// message, key 순으로 메세지 몸체를 파싱
// filter, textReplacer 로 커스텀 가능
func (j *job) makeBody(rowMapList []map[string]string) (string, bool, error) {

	var body string

	filteredRows := 0

	for _, rowMap := range rowMapList {

		var line string

		if j.filter != nil {
			if passed, err := j.filter(rowMap); err != nil {
				return "", false, err
			} else if !passed {
				filteredRows++
			}
		}

		var replaceErr error
		if j.textReplacer != nil {
			rowMap, replaceErr = j.textReplacer(rowMap)
			if replaceErr != nil {
				return "", false, replaceErr
			}
		}

		for i, message := range j.messages {
			line += message
			if rowMap[j.columns[i]] != "" {
				line += rowMap[j.columns[i]] + " "
			}
		}
		body += line + "\n"
	}

	if filteredRows == len(rowMapList) {
		return "", false, nil
	}
	return body, true, nil
}

// returns registered job List
func GetJobList() []*job {
	return jobLists
}

func SetDefaultToken(token string) {
	defaultToken = token
}
