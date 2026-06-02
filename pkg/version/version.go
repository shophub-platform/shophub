// Package version drži verziju build-a (override-uje se preko -ldflags pri build-u).
package version

// Version je semantička verzija aplikacije. Postavlja se pri build-u:
//
//	go build -ldflags "-X github.com/shophub-platform/shophub/pkg/version.Version=v1.2.3"
var Version = "0.0.0-dev"
