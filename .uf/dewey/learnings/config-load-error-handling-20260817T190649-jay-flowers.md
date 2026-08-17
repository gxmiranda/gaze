---
tag: config-load-error-handling
author: jay-flowers
category: context
created_at: 2026-08-17T19:06:49Z
identity: config-load-error-handling-20260817T190649-jay-flowers
tier: draft
---

In config.LoadFromDir, the `config.Load` function already handles `os.IsNotExist` at line 166-168 by returning `(DefaultConfig(), nil)`. This means any error that reaches `LoadFromDir`'s `if err != nil` branch is always a real error (YAML parse failure, validation failure, or permission error) — never file-not-found. No `os.IsNotExist` check is needed in `LoadFromDir` itself. This was verified during the config-load-warning implementation (design decision D2).
