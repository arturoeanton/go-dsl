module github.com/arturoeanton/go-dsl/examples/image_dsl

go 1.24.5

require (
	github.com/arturoeanton/go-dsl v0.0.0
	golang.org/x/image v0.14.0
)

require gopkg.in/yaml.v3 v3.0.1 // indirect

replace github.com/arturoeanton/go-dsl => ../..
