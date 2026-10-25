# Develop the documentation locally

The website is an independent npm package under `website/`. Use the Node patch in `website/.node-version` and dependencies from `package-lock.json`.

## Start the development server

```sh
cd website
npm ci
npm run dev
```

Open the localhost URL printed by Astro. The supervisor watches canonical documentation, mapped examples, reference inputs and presentation files. Stop it with Ctrl-C before starting another server. Use HTTP: generated HTML opened with `file://` cannot provide working navigation, modules and search.

## Check the built site

With the development server stopped:

```sh
npm run check
npm run references:check
npm test
npm run build
npm run preview
```

Preview serves the static site and Pagefind index at `http://127.0.0.1:4321/`. Test search here; the development server has no production search index. Local navigation should stay on localhost.

Stop preview with Ctrl-C before running browser checks, which start their own server on port 4321. Install Chromium once after dependency installation, and again when Playwright changes:

```sh
npm run test:browsers:install
npm run test:e2e
npm run ci:verify-disabled
```

Alternatively, `npm run verify:website` runs the website checks, one build and browser tests together; Chromium must already be installed and port 4321 free. See the [website README](../../website/README.md) for stages and browser-cache settings.

## Troubleshoot the right layer

Dependency and browser installation may need network access; content builds use local inputs. For a missing browser, run `npm run test:browsers:install`. Keep Go/Rust pins unchanged when resolving Node problems.

For missing routes, inspect the content map. For broken fragments, check the target heading. For stale example fingerprints, review the source and evidence before regeneration. Source archives remain usable: absent Git metadata produces unknown commit identity and repository-relative source attribution.

## Keep native work opt-in

Website commands do not run the compiler or native suites. Compiler bootstrap and expensive validation remain subject to the [temporary validation pause](../development-harness.md#temporary-validation-pause).

The optional example verifier requires explicit invocation and an existing compiler; it does not replace acceptance. Website CI/CD remains prepared but disabled. No deployment has been performed.
