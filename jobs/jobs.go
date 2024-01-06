package jobs

import (
	"connectionPool"
	"log"
	"sendMessage"
	"util"

	"github.com/go-co-op/gocron/v2"
)

var defaultToken string

var jobLists = []*job{}

type job struct {
	name         string
	description  string
	query        string
	messages     []string
	columns      []string
	cron         string
	head         string
	filter       jobFilter
	recentStatus bool
	token        string
}

type Job interface {
	Register(s gocron.Scheduler)
	Execute()
	GetDescription() string
	GetName() string
	GetCron() string
	GetRecentStatus() bool
	SetRecentStatus(bool)
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
	SetDescription(string) JobBuilder
	Build() *job
}

type jobFilter func(map[string]string) bool

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

// setter (optional) : description
// 해당 job 에 대한 간략한 설명
func (jb *jobBuilder) SetDescription(description string) JobBuilder {
	jb.job.description = description
	return jb
}

// returns builded job
func (jb *jobBuilder) Build() *job {
	//validation -> name,cron
	jb.job.SetRecentStatus(true)
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
		j.SetRecentStatus(false)
		return
	}

	head := j.head
	body, pass := makeBody(rowMapList, j.messages, j.columns, j.filter)
	if !pass {
		return
	}

	token := j.token

	//Line API Token
	if token == "" {
		token = defaultToken
	}

	err = sendMessage.SendLineMessage(head+body, token)
	if err != nil {
		log.Println("error while sending Msg : ", err)
	}
	j.SetRecentStatus(true)
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
func (j *job) GetRecentStatus() bool {
	return j.recentStatus
}

// set job.RenetStatus
func (j *job) SetRecentStatus(status bool) {
	j.recentStatus = status
}

// push job to job list
func PushJob(j *job) {
	jobLists = append(jobLists, j)
}

// message, key 순으로 메세지 몸체를 파싱
func makeBody(rowMapList []map[string]string, messages []string, keys []string, filter jobFilter) (string, bool) {

	var body string

	filteredRows := 0

	for _, rowMap := range rowMapList {

		var line string

		if filter != nil {
			if !filter(rowMap) {
				filteredRows++
			}
		}

		for i, message := range messages {
			line += message
			if rowMap[keys[i]] != "" {
				line += rowMap[keys[i]] + " "
			}
		}
		body += line + "\n"
	}

	if filteredRows == len(rowMapList) {
		return "", false
	}
	return body, true
}

// returns registered job List
func GetJobList() []*job {
	return jobLists
}

func SetDefaultToken(token string) {
	defaultToken = token
}
