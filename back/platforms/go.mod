module github.com/mbatimel/AMC/platforms

go 1.25.0

require (
	github.com/gofiber/adaptor/v2 v2.2.1
	github.com/gofiber/fiber/v2 v2.52.13
	github.com/google/uuid v1.6.0
	github.com/mbatimel/AMC/access v0.0.0-00010101000000-000000000000
	github.com/mbatimel/AMC/products v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.23.2
	github.com/rs/zerolog v1.35.1
	github.com/valyala/fasthttp v1.72.0
)

require (
	github.com/andybalholm/brotli v1.2.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/klauspost/compress v1.18.6 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.23 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/sys v0.46.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace github.com/mbatimel/AMC/access => ../access

replace github.com/mbatimel/AMC/products => ../products
