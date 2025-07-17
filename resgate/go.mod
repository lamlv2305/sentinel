module github.com/lamlv2305/sentinel/resgate

go 1.23.3

require (
	github.com/casbin/casbin/v2 v2.109.0
	github.com/goccy/go-json v0.10.5
	github.com/google/uuid v1.6.0
	github.com/lamlv2305/sentinel/types v0.0.0-00010101000000-000000000000
)

require (
	github.com/bmatcuk/doublestar/v4 v4.6.1 // indirect
	github.com/casbin/govaluate v1.3.0 // indirect
)

replace github.com/lamlv2305/sentinel/types => ../types
