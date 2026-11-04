/**
 * Builds site/content from the repository documentation, which stays the single
 * source of truth: docs/*.md, docs/trading/*.md, docs/adrs, docs/sdd and the
 * root community files. Run `bun run sync:docs` after editing any of them and
 * commit the result. Writes content/docs/**\/*.mdx and content/navigation.ts.
 */
import fs from "node:fs";
import path from "node:path";

const siteDir = path.join(import.meta.dirname, "..");
const repoDir = path.join(siteDir, "..");
const outDocs = path.join(siteDir, "content", "docs");

type Page = { src: string; slug: string; title?: string; description?: string };
type Category = { title: string; dir: string; prefix?: string; pages: Page[] };

const adrFiles = fs
	.readdirSync(path.join(repoDir, "docs", "adrs"))
	.filter((f) => /^\d{4}-.*\.md$/.test(f))
	.sort();
const sddFiles = fs
	.readdirSync(path.join(repoDir, "docs", "sdd"))
	.filter((f) => /^\d{2}-.*\.md$/.test(f))
	.sort();
const researchFiles = [
	"lean",
	"jesse",
	"freqtrade",
	"octobot",
	"hummingbot",
	"etape",
	"openbb",
	"openstock",
	"openterminal",
	"quantdinger",
];

const categories: Category[] = [
	{
		title: "Getting started",
		dir: "getting-started",
		prefix: "",
		pages: [
			{ src: "docs/product.md", slug: "introduction", title: "Introduction" },
			{ src: "docs/setup-local-with-docker.md", slug: "setup-local-with-docker" },
			{ src: "docs/setup-vps-with-docker.md", slug: "setup-vps-with-docker" },
			{ src: "docs/setup-in-railway.md", slug: "setup-in-railway" },
		],
	},
	{
		title: "Architecture",
		dir: "architecture",
		pages: [
			{ src: "docs/architecture.md", slug: "overview" },
			{ src: "docs/data-model.md", slug: "data-model" },
			{ src: "docs/validation.md", slug: "validation" },
			{ src: "docs/sources.md", slug: "data-sources" },
		],
	},
	{
		title: "Trading knowledge",
		dir: "trading",
		pages: [
			{ src: "docs/trading/README.md", slug: "overview", title: "Overview" },
			{ src: "docs/trading/backtesting.md", slug: "backtesting" },
			{ src: "docs/trading/risk-and-execution.md", slug: "risk-and-execution" },
			{ src: "docs/trading/strategies-and-markets.md", slug: "strategies-and-markets" },
			{ src: "docs/trading/terminal-and-agents.md", slug: "terminal-and-agents" },
		],
	},
	{
		title: "Reference projects",
		dir: "reference-projects",
		pages: researchFiles.map((n) => ({ src: `docs/trading/${n}.md`, slug: n })),
	},
	{
		title: "Decisions",
		dir: "adrs",
		pages: [
			{ src: "docs/adrs/README.md", slug: "index", title: "Architecture decision records" },
			...adrFiles
				.filter((f) => f !== "README.md")
				.map((f) => ({ src: `docs/adrs/${f}`, slug: f.replace(/\.md$/, "") })),
		],
	},
	{
		title: "Specifications",
		dir: "sdd",
		pages: [
			{ src: "docs/sdd/README.md", slug: "index", title: "Design specifications" },
			...sddFiles.map((f) => ({ src: `docs/sdd/${f}`, slug: f.replace(/\.md$/, "") })),
		],
	},
	{
		title: "Operations",
		dir: "operations",
		pages: [
			{ src: "docs/deploy-checklist.md", slug: "deploy-checklist" },
			{ src: "docs/railway.md", slug: "railway" },
		],
	},
	{
		title: "Runbooks",
		dir: "runbooks",
		pages: [
			{ src: "docs/runbooks/README.md", slug: "index", title: "Runbooks" },
			{ src: "docs/runbooks/incident.md", slug: "incident" },
			{ src: "docs/runbooks/restore.md", slug: "restore" },
			{ src: "docs/runbooks/restore-rehearsal-2026-11-04.md", slug: "restore-rehearsal" },
			{ src: "docs/runbooks/key-rotation.md", slug: "key-rotation" },
		],
	},
	{
		title: "Project",
		dir: "project",
		pages: [
			{ src: "docs/brainstorming.md", slug: "brainstorming" },
			{ src: "docs/planning.md", slug: "planning" },
			{ src: "docs/grilling.md", slug: "grilling" },
			{ src: "docs/web-research.md", slug: "web-research" },
			{ src: "docs/dependency-policy.md", slug: "dependency-policy" },
			{ src: "docs/mcp.md", slug: "mcp" },
			{ src: "CONTRIBUTING.md", slug: "contributing" },
			{ src: "SECURITY.md", slug: "security" },
			{ src: "CODE_OF_CONDUCT.md", slug: "code-of-conduct" },
		],
	},
];

const existing = categories.flatMap((c) =>
	c.pages.filter((p) => fs.existsSync(path.join(repoDir, p.src))).map((p) => ({ ...p, cat: c })),
);
const hrefBySource = new Map(
	existing.map((p) => [p.src, `/docs/${p.cat.prefix ?? p.cat.dir}/${p.slug}`.replace("//", "/")]),
);

/** Splits a document into its H1, first paragraph and body without the H1. */
function split(raw: string): { title: string; description: string; body: string } {
	const lines = raw.replace(/\r\n/g, "\n").split("\n");
	let title = "";
	const rest: string[] = [];
	for (const line of lines) {
		if (!title && /^# /.test(line)) {
			title = line.slice(2).trim();
			continue;
		}
		rest.push(line);
	}
	const body = rest.join("\n").replace(/^\s+/, "");
	const para = body.split(/\n\s*\n/).find((p) => p.trim() && !/^(#|[-*>|`]|\d+\.)/.test(p.trim()));
	const plain = (para ?? "")
		.replace(/\[([^\]]+)\]\([^)]*\)/g, "$1")
		.replace(/[`*_<>]/g, "")
		.replace(/\s+/g, " ")
		.trim();
	const description = plain.length > 200 ? `${plain.slice(0, 197).trimEnd()}...` : plain;
	return { title, description, body };
}

/** Makes Markdown safe for MDX and rewrites repository links to site links. */
function toMdx(body: string, from: string): string {
	const out: string[] = [];
	let fence = false;
	for (const line of body.split("\n")) {
		if (/^\s*```/.test(line)) {
			fence = !fence;
			out.push(line);
			continue;
		}
		if (fence) {
			out.push(line);
			continue;
		}
		const parts = line.split(/(`[^`]*`)/);
		out.push(
			parts
				.map((part, i) => {
					if (i % 2 === 1) return part;
					return part
						.replace(/<(https?:\/\/[^>\s]+)>/g, "[$1]($1)")
						.replace(/\]\(([^)#]+\.md)(#[^)]*)?\)/g, (m, target: string, hash?: string) => {
							const abs = path.posix.normalize(path.posix.join(path.posix.dirname(from), target));
							const href = hrefBySource.get(abs);
							return href ? `](${href}${hash ?? ""})` : `](https://github.com/AlexGalhardo/vukem-trading/blob/main/${abs}${hash ?? ""})`;
						})
						.replace(/<!--[\s\S]*?-->/g, "")
						.replace(/\{/g, "&#123;")
						.replace(/\}/g, "&#125;")
						.replace(/<(?!\/?(a|br|code|kbd|sub|sup|strong|em)\b)/g, "&lt;");
				})
				.join(""),
		);
	}
	return out.join("\n");
}

fs.rmSync(outDocs, { recursive: true, force: true });
for (const p of existing) {
	const raw = fs.readFileSync(path.join(repoDir, p.src), "utf8");
	const { title, description, body } = split(raw);
	const fm = [
		"---",
		`title: ${JSON.stringify(p.title ?? title ?? p.slug)}`,
		`description: ${JSON.stringify(p.description ?? (description || (p.title ?? title)))}`,
		"---",
		"",
	].join("\n");
	const dest = path.join(outDocs, p.cat.dir, `${p.slug}.mdx`);
	fs.mkdirSync(path.dirname(dest), { recursive: true });
	fs.writeFileSync(dest, `${fm}${toMdx(body, p.src).trimEnd()}\n`);
}

const nav = categories
	.map((c) => ({ ...c, pages: c.pages.filter((p) => existing.some((e) => e.src === p.src)) }))
	.filter((c) => c.pages.length > 0)
	.map((c) => {
		const prefix = c.prefix === undefined ? "" : `    prefix: ${JSON.stringify(c.prefix)},\n`;
		return `  {\n    title: ${JSON.stringify(c.title)},\n    dir: ${JSON.stringify(c.dir)},\n${prefix}    pages: [${c.pages.map((p) => JSON.stringify(p.slug)).join(", ")}],\n  },`;
	})
	.join("\n");
fs.writeFileSync(
	path.join(siteDir, "content", "navigation.ts"),
	`/**
 * Generated by scripts/sync-docs.ts from the repository documentation. Do not edit by hand:
 * change the source Markdown and run \`bun run sync:docs\`.
 */
export type NavCategory = {
  title: string;
  dir: string;
  prefix?: string;
  pages: string[];
};

export const navigation: NavCategory[] = [
${nav}
];
`,
);
console.log(`synced ${existing.length} pages`);
