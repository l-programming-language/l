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

The aggregate checks release policy, references, content/types, unit tests, one static build, and browser behavior. It does not run the L compiler. Browser checks manage their own preview server and use `.cache/ms-playwright` by default. `PLAYWRIGHT_BROWSERS_PATH` can override that cache location.

After a successful build, inspect the site with:

```sh
npm run preview
```

Preview serves `http://127.0.0.1:4321/`, including the Pagefind search index that is absent from the development server. Use HTTP rather than opening HTML with `file://`. Stop preview before another verification run. Individual commands in [package.json](../../website/package.json) are available for focused diagnosis.

## Troubleshoot the right layer

- Missing dependencies or browser: repeat `npm ci` or `npm run test:browsers:install` with the pinned Node. These downloads need network access. Content builds use local inputs.
- Missing route or fragment: check the content-map entry or target heading.
- Stale example or reference: review its source and evidence before regenerating snapshots. See [authoring](documentation.md).
- No Git metadata: source archives still build, with unknown commit identity and repository-relative source attribution. `npm run test:archive` checks this path.

## Keep native work opt-in

Compiler bootstrap and native validation follow the [validation policy](../development-harness.md#validation-authorization). Task authorization persists through its checks. The optional example verifier requires authorization and an existing compiler with its native prerequisites:

```sh
npm run examples:verify -- --compiler /absolute/path/to/l --id basic
```

Omit `--id` to check every complete registered example. This does not execute
arbitrary Markdown fences or replace the temporary hello-world gate.
`website/reports/examples-verification.json` is replaced on each invocation.
`running` is incomplete, `failed` names the failed stage, and only `passed` records
a completed selection.

### Test the installation flow

The separate `test:install` command also requires explicit authorization. It downloads a real GitHub release and its Rust tools, then builds and runs the canonical [basic program](../../demo/basic/main.l). It is excluded from routine `verify:website` and documentation CI.

Use a supported native host with C/linker tools already installed. The release named by root `VERSION` must already be published. First build the current website and keep its preview running:

```sh
npm run build
npm run preview
```

In another terminal, from `website/`, run:

```sh
npm run test:install
```

The check requires `http://127.0.0.1:4321/install.sh` to serve the exact bootstrap from the current build. Only that local preview request uses HTTP. Release and Rust downloads use HTTPS. It installs into a disposable home with separate Cargo, Rustup, and native-cache directories, using a custom executable location. It checks immediate activation and fresh Bash/Zsh sessions, then compiles the unchanged canonical program offline and checks its exact output. Linux hosts without Zsh record that shell check as unavailable.

It then runs the documented versioned installer and manual PATH commands in a separate home with a fresh native cache, reusing the isolated Rust dependencies prepared above. This stage compares the installed compiler bytes, checks that no bootstrap activation file was created, and separately builds and runs the same program offline. Its dependency setup is warm; the original bootstrap and shell observations remain separate.

Per-stage logs and `report.json` are retained under `website/reports/install-smoke-*/`. Temporary installation files are removed afterward. Report the host, released version, completed stages, unavailable shell checks, and evidence path. This verifies the installation flow on that host. It is not the temporary hello-world gate or evidence for other platforms.

Website checks are outside the temporary gate and are not implied by a documentation edit. The legacy-named `ci:verify-disabled` command validates the canonical workflows and their authorization boundaries. Publication is disabled pending separate reviewed [activation](../../website/DEPLOYMENT.md#deliberate-activation-after-review).
