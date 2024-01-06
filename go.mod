module main

go 1.21.3

require github.com/go-sql-driver/mysql v1.7.1 //indirect

require connectionPool v0.0.0

require (
	github.com/go-co-op/gocron/v2 v2.1.2
	github.com/robfig/cron/v3 v3.0.1 // indirect
)

require (
	github.com/google/uuid v1.5.0 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	golang.org/x/exp v0.0.0-20231219180239-dc181d75b848 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace sendMessage => ./sendMessage

replace connectionPool => ./connectionPool

replace util => ./util
