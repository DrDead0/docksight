## Summary

<!-- What does this change do, and why does it exist? -->

## Type of change

- [ ] Bug fix
- [ ] New feature
- [ ] Refactor / cleanup
- [ ] Docs
- [ ] Infrastructure / CI
- [ ] Breaking change

## Areas touched

- [ ] `apps/server` (NestJS)
- [ ] `apps/web` (React / Vite)
- [ ] `apps/agent` (Go)
- [ ] `packages/protocol` (shared WebSocket contracts)
- [ ] `infrastructure` / `.github/workflows`
- [ ] `docs`

## Related issues

 Closes <!-- #123 -->

## How to test

<!-- Steps a reviewer can follow to verify the change. -->

1.
2.

## Checklist

- [ ] Change is focused and stays within the existing modular monolith boundaries.
- [ ] Ran the relevant checks locally (see below).
- [ ] Updated docs under `docs/` if module boundaries or the protocol changed.
- [ ] Added an ADR under `docs/decisions/` if this is a larger architectural change.

<details>
<summary>Local checks</summary>

```bash
# packages/protocol must be built first — apps/server and apps/web
# import it from its dist output.
npm run build --workspace=@docksight/protocol
npm run test  --workspace=@docksight/protocol

npm run build --workspace=@docksight/server
npm run lint  --workspace=@docksight/server
npm run test  --workspace=@docksight/server

npm run build --workspace=@docksight/web
npm run lint  --workspace=@docksight/web
```

</details>

## Protocol changes

<!-- Delete if packages/protocol is untouched. -->

- [ ] Updated `packages/protocol/src` **and** the fixtures under `packages/protocol/fixtures`.
- [ ] Server, web, and agent were all updated for the new contract.
- [ ] Noted whether older agents remain compatible.

## Screenshots

<!-- UI changes only. -->
