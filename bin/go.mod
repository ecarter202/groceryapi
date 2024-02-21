module groceryAPI

go 1.22.0

replace (
	api => ./api
	grocery => ../grocery
)

require (
	github.com/gocraft/web v0.0.0-20190207150652-9707327fb69b
	grocery v0.0.0-00010101000000-000000000000
)

require (
	api v0.0.0-00010101000000-000000000000 // indirect
	github.com/kevinburke/go.uuid v1.2.0 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)
