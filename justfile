binary := "nemesis"
bin_dir := "bin"

# list available recipes
default:
    @just --list

# compile the binary into ./bin
[group('build')]
build:
    go build -o {{bin_dir}}/{{binary}} ./cmd/nemesis

# run all tests
[group('test')]
test:
    go test ./...

# run golangci-lint
[group('dev')]
lint:
    golangci-lint run

# format the code
[group('dev')]
fmt:
    golangci-lint fmt

# tidy module dependencies
[group('dev')]
tidy:
    go mod tidy

# full gate: lint + test + build. all must pass before committing
[group('dev')]
ci: lint test build

# build and run
[group('run')]
run:
    go run ./cmd/nemesis

# build and run with the AI visualiser window
[group('run')]
run-visualiser:
    go run ./cmd/nemesis --visualiser

# benchmark the renderer (Frame cost)
[group('test')]
bench:
    go test -run=^$ -bench=. -benchmem ./internal/render

# fuzz the world generator for a fixed time
[group('test')]
fuzz:
    go test -run=^$ -fuzz=FuzzGenerate -fuzztime=30s ./internal/world

# regenerate the documentation demo media
[group('run')]
demos:
    # First-person clip and stills render headlessly through the software renderer.
    go run ./tools/demogen
    # The AI-visualiser demo records the real window (wrap in xvfb-run when headless).
    go run ./cmd/nemesis --record docs/demos/hunter.gif --seed 42 --record-frames 140

# run govulncheck
[group('dev')]
vulncheck:
    govulncheck ./...

# remove build artifacts
[group('dev')]
clean:
    rm -rf {{bin_dir}} dist
