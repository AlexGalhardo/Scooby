# Framework feedback rule

This project is a real-world test of the Vukem framework. Every Vukem bug, missing feature, awkward API, security issue or wrong or missing documentation page you hit gets one entry in the shared `vukem-feedback.md` at the workspace root (`C:\dev\hexascore-appukem-feedback.md`, outside every repo; per-repo copies are legacy). The vukem maintainer fixes entries on the release PR and updates their `Status`.

- Entry: `## TR-<n>: short title` (next free number for the `TR-` prefix; other projects use `VE-`, `HS-`, `ST-`) with the fields `Severity`, `Date`, `Docs`, `Finding`, `Minimal repro`, `Suggested fix`, `Status` (`open` until the maintainer sets it).
- Cite the docs you consulted with <https://vukem.dev/docs> URLs. Mark gaps `[SITE PAGE MISSING]` (the page does not exist) or `[NOT IN VUKEM DOCS]` (the page exists, the API or behavior is not there). Never invent a Vukem API.
- Apply a local workaround and mark it in code with `// workaround(vukem): TR-<n>`, then keep going.
- Never edit a Vukem framework checkout from this repository.
