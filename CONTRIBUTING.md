# Contributing

`wapi` is a library, not a framework: the caller composes it. A constructor
that discovers things by reflection and calls back into application code does
not belong here.

Two rules that shape what is accepted:

1. **Nothing here knows about a domain.** No users, no orders, no tenders.
   If a change needs to know what the data means, it belongs in the
   application.

2. **Dependencies are part of the public interface.** `go.mod` lists what
   every adopter inherits. Adding to it is a deliberate decision, discussed in
   the issue before the pull request.

## Developing alongside an application

Point a local workspace at your checkout instead of the published version:

```bash
# in the application's repository, go.work is gitignored
go work init .
go work use ../wapi
```

## Before opening a pull request

```bash
gofmt -w .
go mod tidy
go vet ./...
go test -race ./...
```
