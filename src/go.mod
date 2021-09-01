module github.com/sufficit/sufficit-quepasa-fork

require (
	github.com/go-chi/chi v4.1.1+incompatible
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/golang-migrate/migrate/v4 v4.14.1 // indirect
	github.com/joho/godotenv v1.3.0 // indirect
	github.com/prometheus/client_golang v1.6.0
	github.com/sufficit/sufficit-quepasa-fork/controllers v0.0.0
	github.com/sufficit/sufficit-quepasa-fork/models v0.0.0
	github.com/sufficit/sufficit-quepasa-fork/library v0.0.0
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

replace github.com/sufficit/sufficit-quepasa-fork/models => ./models
replace github.com/sufficit/sufficit-quepasa-fork/library => ./library
replace github.com/sufficit/sufficit-quepasa-fork/controllers => ./controllers
replace github.com/sufficit/sufficit-quepasa-fork => ./

replace github.com/Rhymen/go-whatsapp => github.com/sufficit/sufficit-go-whatsapp v0.1.11

go 1.14
