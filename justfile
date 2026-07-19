binary := "nemesis"
bin_dir := "bin"

default: build

# Compile the binary into ./bin
build:
    go build -o {{bin_dir}}/{{binary}} ./cmd/nemesis

# Build and run
run:
    go run ./cmd/nemesis

# Build and run with the AI visualiser window
run-visualiser:
    go run ./cmd/nemesis --visualiser

# Run all tests
test:
    go test ./...

# Run go vet
vet:
    go vet ./...

# Run golangci-lint
lint:
    golangci-lint run

# Benchmark the renderer (Frame cost)
bench:
    go test -run=^$ -bench=. -benchmem ./internal/render

# Fuzz the world generator for a fixed time
fuzz:
    go test -run=^$ -fuzz=FuzzGenerate -fuzztime=30s ./internal/world

# Run govulncheck
vulncheck:
    govulncheck ./...

# Tidy module dependencies
tidy:
    go mod tidy

# Remove build artifacts
clean:
    rm -rf {{bin_dir}} dist
