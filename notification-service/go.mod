module notification-service

go 1.21

require (
	github.com/google/uuid v1.6.0
	github.com/rs/zerolog v1.31.0
	github.com/sendgrid/sendgrid-go v3.14.0+incompatible
	github.com/twilio/twilio-go v1.15.2
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
)

require (
	github.com/go-chi/chi/v5 v5.2.3 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sendgrid/rest v2.6.9+incompatible // indirect
	github.com/stretchr/testify v1.8.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
)

replace salon-shared => ../salon-shared
