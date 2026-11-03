# Contributing to Zenc SDK for Go

Thank you for contributing to the **Zenc Go SDK**! This library allows Go developers to interact with Zenc Server, publish/subscribe to WebRTC media tracks, and build streaming services and bots in Go.

---

## Development Workflow

### Prerequisites
- **Go 1.22+** (Go 1.23 recommended)

### Setup
1. Fork and clone the repository:
   ```bash
   git clone https://github.com/<your-username>/zenc-sdk-go.git
   cd zenc-sdk-go
   ```
2. Download dependencies:
   ```bash
   go mod download
   ```
3. Run tests:
   ```bash
   go test -v -race ./...
   ```
4. Format your code:
   ```bash
   go fmt ./...
   ```

---

## Guidelines

1. **Idiomatic Go**: Keep API design clean, thread-safe, and context-aware (`context.Context` for cancellation and timeouts).
2. **Commit Conventions**: Use Conventional Commits (`feat:`, `fix:`, `docs:`, `perf:`).
3. **Sign-off commits**: Use `git commit -s` (Developer Certificate of Origin).
4. By submitting a Pull Request, you agree that your contributions will be licensed under the **Apache License 2.0**.

