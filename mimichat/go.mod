module mimichat

go 1.12

require (
	github.com/linuxdeepin/go-gir v0.0.0-20180122081704-5247e6ec0d59
	github.com/rocket049/gettext-go v0.0.0-20190404080233-af421a50b332
	golang.org/x/net v0.0.0-00010101000000-000000000000
	gopkg.in/sorcix/irc.v2 v2.0.0-20190306112350-8d7a73540b90
)

replace golang.org/x/net => github.com/golang/net v0.0.0-20190404232315-eb5bcb51f2a3

replace golang.org/x/text => github.com/golang/text v0.3.0

replace golang.org/x/crypto => github.com/golang/crypto v0.0.0-20190404164418-38d8ce5564a5

replace golang.org/x/sys => github.com/golang/sys v0.0.0-20190403152447-81d4e9dc473e
