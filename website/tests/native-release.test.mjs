// Release control and installer fixture checks are part of ordinary website CI.
// They make no network requests and execute no Go/Rust compiler or real installer.
import '../../scripts/release-tests/github.test.mjs';
import '../../scripts/release-tests/installer.test.mjs';
import '../../scripts/release-tests/package.test.mjs';
