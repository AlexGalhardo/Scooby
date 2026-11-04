# Framework feedback rule

This project is a real-world test of the Vukem framework. Every Vukem bug, missing feature, awkward API, security issue or wrong or missing documentation page you hit gets one entry in `vukem-feedback.md` at the repository root. The file is git-ignored; the owner forwards the useful entries upstream.

- Entry title: `### YYYY-MM-DD - short title`; then what happened, why it matters, and the suggested change.
- Cite the docs you consulted with <https://vukem.dev/docs> URLs. Mark gaps `[SITE PAGE MISSING]` (the page does not exist) or `[NOT IN VUKEM DOCS]` (the page exists, the API or behavior is not there). Never invent a Vukem API.
- Apply a local workaround and mark it in code with `// workaround(vukem): <entry title>`, then keep going.
- Never edit a Vukem framework checkout from this repository.
