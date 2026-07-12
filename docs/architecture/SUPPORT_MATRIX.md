# Foundation Support Matrix

**Status:** Active foundation evidence  
**Scope:** Repository governance and documentation validation only

The lifecycle engine and plugin do not exist yet. This matrix proves the current public
repository foundation on native runners; it must not be presented as product-runtime,
sensitive-data-route, or regulatory support.

| Environment | Foundation validation | Evidence |
|---|---|---|
| Ubuntu, current GitHub-hosted image | Required | Native matrix job |
| macOS, current GitHub-hosted image | Required | Native matrix job |
| Windows, current GitHub-hosted image | Required | Native matrix job |
| Python 3.12 | Pinned in CI | `actions/setup-python` by immutable commit |

Every native job runs the same semantic commands:

```text
python -m unittest discover -s tests -p "test_*.py"
python scripts/validate_docs.py
```

The validator uses only the Python standard library. GitHub issue forms and the label
manifest are JSON-compatible YAML so their structure parses identically without an
unpinned package installation.

## Currently verified invariants

- Required public, agent, decision, template, manifest, workflow, and support files exist.
- Local Markdown links remain inside the repository and resolve.
- D1–D14 have unique stable decision identities, targets, and substantive source anchors.
- Issue forms parse and reference known labels with unique body IDs.
- The label manifest has unique names, valid colors, and descriptions.
- Workflow actions are pinned, all three native runners are present, and no explicit
  platform shell is required.

## Deferred support decisions

The engine implementation issue will publish exact minimum OS versions, CPU
architectures, filesystems, installer/package behavior, Codex client compatibility, and
external runtime requirements. Until those are implemented and tested, only the
foundation scope above is supported.
