Mew: tiny local-first snapshot tool

Usage

Install dependencies and build (requires Go):

```bash
go install ./...
go build -o mew
```

Examples

- Create a snapshot: `mew snap "My snapshot title"`
- Restore a snapshot by id or title prefix: `mew wind <id|title>`

Snapshots and metadata are stored under `.mew/` (immutable archives in `.mew/snaps/`).
