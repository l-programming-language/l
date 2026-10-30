# Develop the documentation locally

Use the Node version in [website/.node-version](../../website/.node-version) and its bundled npm. The website is an independent package with locked dependencies.

## Start the development server

```sh
cd website
npm ci
npm run dev
```

Open the localhost URL printed by Astro. Canonical documentation, examples, reference inputs, and presentation files are watched. Stop the server with Ctrl-C before starting another one.

## Check the built site

With other servers on port 4321 stopped, install Chromium once after setup or a Playwright update, then run the website checks:

```sh
npm run test:browsers:install
npm run verify:website
```

The aggregate checks release policy, references, content/types, unit tests, one static build, and browser behavior. It does not run the L compiler. Browser checks manage their own preview server and use `.cache/ms-playwright` by default; `PLAYWRIGHT_BROWSERS_PATH` can override that cache location.

After a successful build, inspect the site with:

```sh
npm run preview
```

Preview serves `http://127.0.0.1:4321/`, including the Pagefind search index that is absent from the development server. Use HTTP rather than opening HTML with `file://`. Stop preview before another verification run. Individual commands in [package.json](../../website/package.json) are available for focused diagnosis.

## Troubleshoot the right layer

- Missing dependencies or browser: repeat `npm ci` or `npm run test:browsers:install` with the pinned Node. These downloads need network access; content builds use local inputs.
- Missing route or fragment: check the content-map entry or target heading.
- Stale example or reference: review its source and evidence before regenerating snapshots. See [authoring](documentation.md).
- No Git metadata: source archives still build, with unknown commit identity and repository-relative source attribution. `npm run test:archive` checks this path.

## Keep native work opt-in

Compiler bootstrap and expensive native validation follow the [temporary validation pause](../development-harness.md#temporary-validation-pause). The optional example verifier requires explicit authorization and an existing compiler with its native prerequisites:

```sh
npm run examples:verify -- --compiler /absolute/path/to/l --id basic
```

This checks approved examples, not arbitrary Markdown fences or the full acceptance gate. Documentation CI runs website checks on pull requests. The legacy-named `ci:verify-disabled` command validates approved manual release paths and publication authorization. Publication requires an authorized manual dispatch from main; see [deployment](../../website/DEPLOYMENT.md) for activation and recovery.
