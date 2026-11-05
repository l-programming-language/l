# Contracts evidence

The [acceptance catalog](../contracts/acceptance-catalog.json) defines requirements;
the [acceptance map](../contracts/acceptance-map.json) identifies their test
witnesses and current limitations. Mapping integrity checks do not execute those
tests or establish native approval. [Contracts support and limits](contracts.md)
describe the supported behavior.

Historical records are retained at commit
`092d8e3674c21fb158d86d7bb929cceaf6c9596d`. These three GitHub records were retrieved
and their bytes verified before the diaries were removed from the current tree:

| Record | Bytes | SHA-256 |
| --- | ---: | --- |
| [Implementation history](https://github.com/l-programming-language/l/blob/092d8e3674c21fb158d86d7bb929cceaf6c9596d/docs/contracts-implementation.md) | 37,731 | `d6accfb3ac9b487f854bd03e7b06dd88d48461b504f5b55941208cded7dac8c0` |
| [Validation records](https://github.com/l-programming-language/l/blob/092d8e3674c21fb158d86d7bb929cceaf6c9596d/docs/contracts-pr20-validation.json) | 7,901 | `7d6b3720b6aa544b65b7d1882153f8a9cbc743943e4f1d9feb7b255d7a91e1a4` |
| [Earlier mapping and command records](https://github.com/l-programming-language/l/blob/092d8e3674c21fb158d86d7bb929cceaf6c9596d/contracts/acceptance-map.json) | 93,264 | `58a2ed047689cb89f508c802438b7d2e61a8775deda2041be5fc2479b872df4b` |

These records preserve successful and failed attempts, platform and tool versions,
source fingerprints, dirty-checkout qualifications and evidence limitations.
Historical passes apply only to their recorded inputs; they are not a pass for
the current checkout. Positive native examples use explicitly test-only
assumptions; production native trust remains empty.

Raw local files remain in ignored `.harness/runs/` and `.harness/pr20-repair/`.
All 23 path-and-digest references in the validation record matched their 19 unique
local files during this review. Those local files are not a publicly accessible
archive. The separately cited original review file
`/private/tmp/pr20-review-validation-k_ybuxl8/validation.json` was unavailable;
its historical narrative must not be treated as retrievable raw evidence.

Hosted evidence for `092d8e3` was also checked. The
[Linux CI artifact](https://github.com/l-programming-language/l/actions/runs/41602762341/artifacts/13215573513)
download matched SHA-256
`735f10e15a4de45bed974ef624258867a5975fa74886a2c43531652c6e71d930`.
The [documentation artifact](https://github.com/l-programming-language/l/actions/runs/41602762466/artifacts/13215372239)
had accessible metadata, but its downloaded bytes were not verified.
Their reported expiry times were respectively **2026-11-10 08:34:22 UTC** and
**2026-11-10 08:25:35 UTC**. These expiring artifacts are not durable archive
guarantees. No evidence was republished or promoted into native approval.
