module github.com/kjanat/slimacademy

go 1.24.4

require gopkg.in/yaml.v3 v3.0.1

// Development tools using Go 1.24 tool directive
tool github.com/golangci/golangci-lint/cmd/golangci-lint
tool honnef.co/go/tools/cmd/staticcheck
tool github.com/securecodewarrior/gosec/v2/cmd/gosec
tool golang.org/x/vuln/cmd/govulncheck
