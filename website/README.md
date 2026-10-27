# L documentation website

This independent npm package builds the canonical repository documentation with Astro and Starlight. The production origin is **https://ldocs.org/**; local navigation stays on localhost. The package's `private: true` flag prevents npm publication.

[Documentation checks](../.github/workflows/docs-check.yml) run on pull requests. Manual release workflows are implemented, but **publication remains disabled**. See [deployment](DEPLOYMENT.md) for activation and recovery procedures.

## Setup and local use

Follow [local development](../docs/contributing/local-development.md) for the pinned Node runtime, installation, development server, and preview commands.

## Website verification

The [verification recipe](../docs/contributing/local-development.md#check-the-built-site) runs content/reference checks, unit tests, one static build, and browser checks. Reports and screenshots are ignored under `reports/`; [recorded results](reports/validation.md) apply to their documented revision.

## Content and references

Edit canonical prose under `docs/`, not generated website projections. [Authoring documentation](../docs/contributing/documentation.md) explains the content map, links, examples, and reference regeneration. [Architecture](ARCHITECTURE.md) describes projection and source ownership.

## Explicitly opt-in language examples

Website commands do not run the compiler or native suites. The optional example verifier follows the [native validation policy](../docs/contributing/local-development.md#keep-native-work-opt-in).

## Troubleshooting and archives

See [troubleshooting](../docs/contributing/local-development.md#troubleshoot-the-right-layer). `npm run test:archive` additionally checks website builds without Git metadata in a temporary source archive; it uses the existing locked dependency tree. Keep caches, generated projections, build output, and test reports out of Git.
