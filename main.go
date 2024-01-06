package main

import (
	"connectionPool"
	"fmt"
	"jobs"
	"log"
	"net/http"
	"os"
	"queries"

	"github.com/go-co-op/gocron/v2"
	"gopkg.in/yaml.v2"
)

type config struct {
	DSN        string `yaml:"DSN"`
	TOKEN_LINE string `yaml:"TOKEN_LINE"`
}

var scheduler gocron.Scheduler

// 모듈 등록
func initJobs() {
	queries.ExampleQuery().Register(scheduler)
}

func main() {

	//설정 등록
	config := readConfig()

	// DB 커넥션풀 초기화
	connectionPool.SetDsn(config.DSN)
	db := connectionPool.GetConnection()
	defer db.Close()

	//라인 토큰 초기화
	jobs.SetDefaultToken(config.TOKEN_LINE)

	// 크론 스케줄러 초기화
	var nserr error
	scheduler, nserr = gocron.NewScheduler()
	if nserr != nil {
		log.Fatal(nserr)
	}
	defer scheduler.Shutdown()

	// jobs 초기화
	initJobs()

	scheduler.Start()

	http.HandleFunc("/", describe)
	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/status", getAllStatus)
	log.Fatal(http.ListenAndServe(":10080", nil))

	select {}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

// jobs lis 조회용 api
func describe(w http.ResponseWriter, r *http.Request) {
	jobList := jobs.GetJobList()
	fmt.Fprintf(w, "===== registered job list ===== \n")

	for _, job := range jobList {
		fmt.Fprintf(w, job.GetName()+" : "+job.GetDescription()+", cron : "+job.GetCron()+"\n")
	}

	fmt.Fprintf(w, "===== registered jobs End ===== \n")
}

// job 실패/성공 상태 조회용 api
func getAllStatus(w http.ResponseWriter, r *http.Request) {
	jobList := jobs.GetJobList()
	fmt.Fprintf(w, "===== job status list===== \n")

	for _, job := range jobList {
		status := "OK"
		if !job.GetRecentStatus() {
			status = "FAILED"
		}
		fmt.Fprintf(w, job.GetName()+" : "+status+"\n")
	}

	fmt.Fprintf(w, "===== job status list End ===== \n")
}

func readConfig() config {
	configFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("설정 파일을 읽어올 수 없습니다: %v", err)
	}

	var cnf config

	err = yaml.Unmarshal(configFile, &cnf)
	if err != nil {
		log.Fatalf("설정 파일을 파싱할 수 없습니다: %v", err)
	}

	return cnf
}
