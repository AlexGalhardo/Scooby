# Commit rules

- Conventional Commits 1.0, English: `type(scope): description`.
- Types: feat, fix, perf, refactor, docs, test, build, ci, chore, style, revert.
- Scope = area, lowercase: `engine`, `risk`, `backtest`, `adapters`, `data`, `strategy`, `web`, `site`, `auth`, `infra`, `deps`, `repo`, `docs`.
- Description: imperative, lowercase start, no trailing period, header at most 100 characters.
- Breaking change: `!` after type or scope plus a `BREAKING CHANGE:` footer.
- Reference the ROADMAP.md task in the body: `Refs: P1.2.3`.
- One logical change per commit. Work on short-lived branches and merge into `main` through a pull request; the PR title is a Conventional Commit.
- The POSIX hook `.githooks/commit-msg` enforces the header (enable it with `git config core.hooksPath .githooks`).
- Subjects say what changed for a user or operator (`fix(risk): reject orders without a stop`), not the task id or the process.
- Release notes are generated from these subjects by git-cliff; `docs`, `test`, `ci`, `build`, `chore` and `style` commits stay out of the notes, so a user-visible change must be `feat`, `fix`, `perf` or `refactor`.
- Never rewrite history or force-push `main`.
