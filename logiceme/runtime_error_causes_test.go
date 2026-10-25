//go:build darwin || linux

package Logiceme

import "testing"

// Independent capture and cleanup failures share deliberately identical text.
// Only their original native cause identities can distinguish them.
func TestRuntimeJoinedNativeCauses(t *testing.T) {
	t.Parallel()
	driver := buildErrorRuntime(t, `mod runtime;
#[path = "process_unix.rs"] mod process;
use std::io::{self, Write};
use std::sync::Arc;
use runtime::test_support::{self, Dependencies, SignalTarget, Stream};
#[derive(Debug)]
struct Cause(Arc<()>);
impl std::fmt::Display for Cause {
    fn fmt(&self, out: &mut std::fmt::Formatter<'_>) -> std::fmt::Result { out.write_str("injected native failure") }
}
impl std::error::Error for Cause {}
struct BrokenCapture(Arc<()>);
impl Write for BrokenCapture {
    fn write(&mut self, _: &[u8]) -> io::Result<usize> { Err(io::Error::other(Cause(self.0.clone()))) }
    fn flush(&mut self) -> io::Result<()> { Ok(()) }
}
struct Probe { capture: Arc<()>, cleanup: Arc<()> }
impl Dependencies for Probe {
    fn capture<'a>(&self, stream: Stream, destination: Box<dyn Write + Send + 'a>) -> Box<dyn Write + Send + 'a> {
        if stream == Stream::Stdout { Box::new(BrokenCapture(self.capture.clone())) } else { destination }
    }
    fn signal(&self, target: SignalTarget, _: nix::sys::signal::Signal) -> io::Result<()> {
        let pid = match target { SignalTarget::Group(pid) => -(pid as i32), SignalTarget::Leader(pid) => pid as i32 };
        match nix::sys::signal::kill(nix::unistd::Pid::from_raw(pid), nix::sys::signal::Signal::SIGKILL) {
            Ok(()) | Err(nix::errno::Errno::ESRCH) => {},
            Err(error) => return Err(io::Error::from_raw_os_error(error as i32)),
        }
        Err(io::Error::other(Cause(self.cleanup.clone())))
    }
}
fn contains(error: &runtime::Error, identity: &Arc<()>) -> bool {
    let mut source = std::error::Error::source(error);
    while let Some(current) = source {
        if current.downcast_ref::<io::Error>()
            .and_then(io::Error::get_ref)
            .and_then(|native| native.downcast_ref::<Cause>())
            .is_some_and(|cause| Arc::ptr_eq(&cause.0, identity)) { return true; }
        source = current.source();
    }
    error.causes().iter().any(|cause| contains(cause, identity))
}
fn main() {
    let executor = tokio::runtime::Builder::new_current_thread().enable_all().build().unwrap();
    executor.block_on(async {
        let probe = Arc::new(Probe { capture: Arc::new(()), cleanup: Arc::new(()) });
        let runtime = test_support::new_observed(probe.clone(), Box::pin(tokio::io::sink()), Box::pin(tokio::io::sink())).unwrap();
        let app = runtime.app("companion/helper");
        app.run().unwrap();
        let invocation = app.wait().await.unwrap();
        let snapshot = invocation.snapshot().unwrap();
        let error = snapshot.error.unwrap();
        assert!(contains(&error, &probe.capture), "capture identity lost");
        assert!(contains(&error, &probe.cleanup), "cleanup identity lost");
        let cloned = error.clone();
        assert!(contains(&cloned, &probe.capture) && contains(&cloned, &probe.cleanup), "clone changed identity");
        assert_eq!(snapshot.status, Some(0), "native successful exit changed");
        let cleanup = runtime.shutdown().await.unwrap_err();
        assert!(contains(&cleanup, &probe.cleanup), "shutdown lost cleanup identity");
        assert!(!contains(&cleanup, &probe.capture), "unrelated capture cause entered cleanup");
    });
}
`, true)
	runtimeExecute(t, driver, nil, nil, 0, "", "")
}
