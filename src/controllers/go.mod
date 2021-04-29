module github.com/sufficit/sufficit-quepasa-fork/controllers

require (
	github.com/Rhymen/go-whatsapp v0.1.1-0.20200429202648-5e33cb4ac551
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dlclark/regexp2 v1.1.6 // indirect
	github.com/go-chi/chi v4.1.1+incompatible
	github.com/go-chi/jwtauth v4.0.4+incompatible
	github.com/gorilla/websocket v1.4.2
	github.com/jmoiron/sqlx v1.2.0
	github.com/lib/pq v1.5.2
	github.com/nbutton23/zxcvbn-go v0.0.0-20180912185939-ae427f1e4c1d
	github.com/prometheus/client_golang v1.6.0
	github.com/trustelem/zxcvbn v1.0.1
)

replace github.com/sufficit/sufficit-quepasa-fork/controllers => ./

go 1.14
