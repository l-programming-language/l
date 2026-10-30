// Native process, stream, synchronization and executor adapters. Language
// lifecycle policy belongs to the composed runtime source dictionary.
use std::future::Future;
use std::io::{self, Write};
use std::pin::Pin;
use std::process::ExitStatus;
use std::sync::{Arc, Mutex, MutexGuard};
use tokio::io::AsyncWrite;
use tokio::process::{Child, Command};
use crate::{process, runtime::Invocation};

pub(crate) type Wait<'a> = Pin<Box<dyn Future<Output = io::Result<ExitStatus>> + Send + 'a>>;
pub(crate) type Output = Pin<Box<dyn AsyncWrite + Send>>;
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub(crate) enum Stream { Stdout, Stderr }
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub(crate) enum Checkpoint { PreStart, SuccessfulStart, CapturedStopTarget, PrePublication }
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub(crate) enum SignalTarget { Group(u32), Leader(u32) }
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub(crate) enum TerminalRequest { Exit, Failure, Signal(u8) }

pub(crate) trait Dependencies: Send + Sync {
    fn start(&self, command: &mut Command) -> io::Result<Child> { command.spawn() }
    fn wait<'a>(&'a self, child: &'a mut Child) -> Wait<'a> { Box::pin(child.wait()) }
    fn signal(&self, target: SignalTarget, _signal: nix::sys::signal::Signal) -> io::Result<()> {
        match target { SignalTarget::Group(pid) => process::terminate_group(pid), SignalTarget::Leader(pid) => process::terminate_leader(pid) }
    }
    fn capture<'a>(&self, _stream: Stream, destination: Box<dyn Write + Send + 'a>) -> Box<dyn Write + Send + 'a> { destination }
    fn checkpoint(&self, _event: Checkpoint, _invocation: Arc<Invocation>) {}
    fn terminal_entered(&self, _request: TerminalRequest) {}
}
pub(crate) struct NativeDependencies;
impl Dependencies for NativeDependencies {}

// Preserve the host lock's retained state after a poisoned acquisition.
pub(crate) fn lock<T>(mutex: &Mutex<T>) -> MutexGuard<'_, T> {
    mutex.lock().unwrap_or_else(std::sync::PoisonError::into_inner)
}

// Wait for readiness in the caller's explicit priority order. The source
// composition decides what each ready result means and performs cleanup.
pub(crate) async fn race<A: Future, B: Future>(first: A, second: B) -> std::result::Result<A::Output, B::Output> {
    tokio::select! { biased; value = first => Ok(value), value = second => Err(value) }
}

// Native trait adaptation delegates all error behavior to authored words.
impl Clone for crate::runtime::Error {
    fn clone(&self) -> Self { crate::runtime::error_copy(self) }
}
impl std::fmt::Debug for crate::runtime::Error {
    fn fmt(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result { crate::runtime::error_display(self, formatter) }
}
impl std::fmt::Display for crate::runtime::Error {
    fn fmt(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result { crate::runtime::error_display(self, formatter) }
}
impl std::error::Error for crate::runtime::Error {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> { crate::runtime::error_native_cause(self) }
}
impl From<io::Error> for crate::runtime::Error {
    fn from(error: io::Error) -> Self { crate::runtime::error_from_io(error) }
}
impl Clone for crate::runtime::Runtime {
    fn clone(&self) -> Self { crate::runtime::runtime_copy(self) }
}
impl Clone for crate::runtime::Snapshot {
    fn clone(&self) -> Self { crate::runtime::snapshot_copy(self) }
}
impl Clone for crate::runtime::Terminal {
    fn clone(&self) -> Self { crate::runtime::terminal_copy(self) }
}

// Preserve the original native panic payload while adapting Future::poll.
pub(crate) async fn catch<F: Future>(future: F) -> std::result::Result<F::Output, Box<dyn std::any::Any + Send>> {
    let mut future = std::pin::pin!(future);
    std::future::poll_fn(|context| match std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| future.as_mut().poll(context))) {
        Ok(std::task::Poll::Pending) => std::task::Poll::Pending,
        Ok(std::task::Poll::Ready(value)) => std::task::Poll::Ready(Ok(value)),
        Err(panic) => std::task::Poll::Ready(Err(panic)),
    }).await
}
