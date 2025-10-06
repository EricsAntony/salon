module payment-service

go 1.22.0

toolchain go1.24.1

require (
	github.com/go-chi/chi/v5 v5.0.11
	github.com/go-chi/cors v1.2.1
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	github.com/razorpay/razorpay-go v1.3.0
	github.com/rs/zerolog v1.32.0
	github.com/stripe/stripe-go/v76 v76.16.0
	salon-shared v0.0.0-00010101000000-000000000000
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	golang.org/x/sys v0.18.0 // indirect
)

replace salon-shared => ../salon-shared
