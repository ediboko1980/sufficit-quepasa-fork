module github.com/sufficit/sufficit-quepasa-fork

require (
	github.com/Rhymen/go-whatsapp v0.1.1-0.20200429202648-5e33cb4ac551 // indirect
	github.com/cosmtrek/air v1.12.1 // indirect
	github.com/creack/pty v1.1.10 // indirect
	github.com/cznic/ql v1.2.0 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/fatih/color v1.9.0 // indirect
	github.com/go-chi/chi v4.1.1+incompatible
	github.com/go-chi/jwtauth v4.0.4+incompatible
	github.com/golang-migrate/migrate/v4 v4.11.0
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/jackc/fake v0.0.0-20150926172116-812a484cc733 // indirect
	github.com/jackc/pgx v3.2.0+incompatible // indirect
	github.com/kshvakov/clickhouse v1.3.5 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/pelletier/go-toml v1.7.0 // indirect
	github.com/prometheus/client_golang v1.6.0
	github.com/sufficit/sufficit-quepasa-fork/controllers v0.0.0
	github.com/sufficit/sufficit-quepasa-fork/models v0.0.0
	golang.org/x/sys v0.0.0-20200513112337-417ce2331b5c // indirect
)

replace github.com/sufficit/sufficit-quepasa-fork => ./

replace github.com/sufficit/sufficit-quepasa-fork/controllers v0.0.0 => ./controllers

replace github.com/sufficit/sufficit-quepasa-fork/models v0.0.0 => ./models

go 1.14
