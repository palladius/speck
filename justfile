
list:
    just -l

# Build the speck binary into bin/.
build:
    go build -o bin/speck ./cmd/speck

# Generate a spec from an idea file, e.g.: just gen etc/webapp-timeliner.md
# Extra args pass through, e.g.: just gen etc/foo.md --model gemini-pro-latest
gen file *args: build
    ./bin/speck oneshot -f {{file}} -o out {{args}}

# Interactively interview, then write a spec. Extra args pass through.
chat *args: build
    ./bin/speck chat -o out {{args}}

# Run the test suite.
test:
    go test ./...

# gofmt, go vet, build, and test — run before committing.
check:
    gofmt -l .
    go vet ./...
    go build ./...
    go test ./...

# Remove build/output artifacts.
clean:
    rm -rf bin out
