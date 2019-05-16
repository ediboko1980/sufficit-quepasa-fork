module gitlab.com/digiresilience/link/quepasa/controllers

require (
	github.com/Rhymen/go-whatsapp v0.0.2-0.20190511164245-5d5100902126
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dlclark/regexp2 v1.1.6 // indirect
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/go-chi/jwtauth v0.0.0-20190109153619-47840abb19b3
	github.com/gorilla/websocket v1.4.0
	github.com/jmoiron/sqlx v1.2.0
	github.com/lib/pq v1.1.1
	github.com/nbutton23/zxcvbn-go v0.0.0-20180912185939-ae427f1e4c1d
	github.com/skip2/go-qrcode v0.0.0-20190110000554-dc11ecdae0a9
	github.com/test-go/testify v1.1.4 // indirect
	github.com/trustelem/zxcvbn v1.0.1
	gitlab.com/digiresilience/link/quepasa/common v0.0.0
	gitlab.com/digiresilience/link/quepasa/models v0.0.0
)

replace gitlab.com/digiresilience/link/quepasa/controllers => ./

replace gitlab.com/digiresilience/link/quepasa/common => ../common

replace gitlab.com/digiresilience/link/quepasa/models => ../models
