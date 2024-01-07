# About
크론 표현식을 사용, 주기적으로 DB에 쿼리를 날리고 라인으로 메세지 보내는 모니터링 프로그램

# How to use
1. config.yaml 파일을 작성하고, DSN(1줄) 과 LINE_TOKEN 항목을 작성
2. queries/example.go 파일을 참고해서 job 파일 작성하고, 사용 시 main 의 initJobs() 함수에서 등록한다.