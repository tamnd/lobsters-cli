---
title: "Quick start"
description: "Run your first lobsters command."
weight: 30
---

Once `lobsters` is on your `PATH`:

```bash
lobsters --help       # see the command tree
lobsters version      # build info
```

This is a fresh scaffold, so the command tree is just `version` for now. Add
your first real command in `cli/`, build on the `lobsters` library package,
and document it here.

A good first command usually fetches one thing and prints it as JSON, so the
output pipes straight into `jq` and the rest of your tools.
