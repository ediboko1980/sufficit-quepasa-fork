module gitlab.com/digiresilience/link/quepasa

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/go-chi/jwtauth v0.0.0-20190109153619-47840abb19b3
	github.com/golang-migrate/migrate/v4 v4.3.1
	gitlab.com/digiresilience/link/quepasa/common v0.0.0
	gitlab.com/digiresilience/link/quepasa/controllers v0.0.0
	gitlab.com/digiresilience/link/quepasa/models v0.0.0
)

replace gitlab.com/digiresilience/link/quepasa => ./

replace gitlab.com/digiresilience/link/quepasa/controllers v0.0.0 => ./controllers

replace gitlab.com/digiresilience/link/quepasa/common v0.0.0 => ./common

replace gitlab.com/digiresilience/link/quepasa/models v0.0.0 => ./models
