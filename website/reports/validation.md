# Website validation — October 15, 2026

Full website verification and the Git-less archive build passed on macOS ARM64
with pinned Node 24.21.0 and npm 11.19.0. Website CI/CD remains disabled; no
language examples, language acceptance suites or deployment commands ran.

## Tested source

The run tested the historical-record cleanup in a dirty working tree based on
`0c942268d2f42ab2f66995a1604c0c167f8ea378`.

- Documentation revision: `sha256:9a9ca19b14b1ff6910ddb92f37b15911b5e90afe8f2d0ddaa8da850f44061db5`.
- Source fingerprint: `sha256:3d5e26391d18b77f80022696c240739bdb635e6ce3ba9a5feeb977eaa1d6c504`.
- Build channel: local/unreleased; generated source links identify the committed base.

## Results

| Check | Result |
| --- | --- |
| `git diff --check` and `scripts/l-harness doctor` | Passed; 493 integrity, toolchain and Markdown checks |
| `npm run verify:website` | All six stages passed |
| Disabled deployment | Three dormant templates checked; no activation |
| References and content | 63 mapped pages, 58 examples and reviewed source fingerprints passed |
| Astro/type check | Zero errors, warnings or hints |
| Unit tests | 69 passed; no failures or skips |
| Static build | 64 routes, 65 HTML files, fresh Pagefind index, valid links and build identity |
| Browser tests | 12 passed; no failures, skips or retries |
| Accessibility | 66 page/viewport/theme observations; zero reported WCAG 2 A/AA or 2.1 AA violations |
| `npm run test:archive` | Real Git-less build in a path with spaces, deterministic preparation, null Git identity and unchanged authored inputs; staging cleaned |

The browser matrix covers eleven pages at 390, 768 and 1440 pixels in both
light and dark themes. Visible lazy images load before capture; eager-image
checks remain. App example bytes, search, keyboard navigation, copy controls,
no-JavaScript reading, genuine 404 responses and local navigation passed.
Visual inspection covered development, Lucid, builtin, documentation-contribution
and release pages at 390 and 1440 pixels in both themes: 20 readable states with
no horizontal document overflow. Search finds the current validation-budget
section and its anchor works. A sweep of 130 projected/built pages found no
references to the deleted documents or retired amendment anchors.

The build reports Astro's existing catch-all/custom-404 priority warning; the
actual 404 route and response pass. The archive reuses installed locked Node
dependencies; it does not claim a fresh dependency installation.

## Integrity and evidence

All 386 protected hashes match. The contract's sixteen required outcomes and
eleven required-absent entries, freeze manifests, compiler/runtime code, fixtures,
expected bytes, locks, assets, language settings, timeouts and concurrency are
unchanged. Both active Service records and the language-validation report remain
byte-for-byte unchanged. Only the builtin reference's 33 reviewed whole-file
fingerprints and their matching generated coverage hashes changed in metadata.
The [language validation](../../testdata/service-execution/validation-2026-10-15.md)
remains separate evidence for its recorded source and host; it was not rerun here.

The complete tracked follow-up is +338/-1148 lines (net -810) against that clean baseline; ignored build output and logs are excluded.

Raw current evidence is ignored under this directory: `history-cleanup-doctor-final.log`,
`history-cleanup-website.log`, `history-cleanup-archive.log`,
`website-verification.json`, `browser-results.json`, `built-output.json`,
`axe-*.json`, `screenshots/history-cleanup/`, `history-cleanup-visuals.json` and
`archive-smoke.json`. Doctor's report is
`.harness/runs/20261015T231315.677580000Z-3662419921/report.json` from the repository root.
