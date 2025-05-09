# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands
- `go get`: Install dependencies
- `go build`: Build the project
- `go test ./...`: Run all tests
- `go test ./path/to/package`: Run tests in a specific package
- `go test ./path/to/package -run TestName`: Run a single test

## Code Style Guidelines
- **Formatting**: Use `gofmt` or `go fmt ./...` to format code
- **Imports**: Group imports into standard library, external packages, and internal packages with blank lines between groups
- **Types**: Use descriptive type names; capitalize exported types
- **Naming**: Use camelCase for variables, PascalCase for exported functions/types
- **Error Handling**: Always check errors; return errors rather than handling them internally
- **Comments**: Document exported functions/types with comments that begin with function/type name
- **Consistency**: Follow existing code patterns, especially for generated code
- **SDKs**: This project is generated from OpenAPI specs; don't modify generated code directly

## Notes
This is a Go SDK for Plex API, generated via Speakeasy. Do not directly modify code; report issues to the original repository.