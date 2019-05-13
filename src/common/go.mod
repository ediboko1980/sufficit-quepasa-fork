module gitlab.com/digiresilience/link/quepasa/common

require (
	github.com/Rhymen/go-whatsapp v0.0.2-0.20190511164245-5d5100902126
	github.com/go-chi/chi v4.0.2+incompatible // indirect
	github.com/go-chi/jwtauth v0.0.0-20190109153619-47840abb19b3
	github.com/google/uuid v1.1.1
	github.com/jmoiron/sqlx v1.2.0
	github.com/lib/pq v1.0.0
	github.com/skip2/go-qrcode v0.0.0-20190110000554-dc11ecdae0a9
	github.com/stretchr/testify v1.3.0 // indirect
	gitlab.com/digiresilience/link/quepasa/models v0.0.0
	golang.org/x/crypto v0.0.0-20190506204251-e1dfcc566284
	golang.org/x/sync v0.0.0-20181221193216-37e7f081c4d4 // indirect
)

replace gitlab.com/digiresilience/link/quepasa/common => ./

replace gitlab.com/digiresilience/link/quepasa/models => ../models
