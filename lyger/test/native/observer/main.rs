//! Test-owned adapter. The generated entry and runtime modules are unchanged.
//! l_entry_with_runtime is an expected private cfg(test) seam, not supplied here.
#[path = "../generated/src/main.rs"]
mod generated;
use generated::runtime;
// Generated runtime's absolute process reference retains its actual implementation.
use generated::process;
use nix::{sys::signal::{self, Signal}, unistd::Pid};
use runtime::test_support::{Checkpoint, Dependencies, SignalTarget, Stream, TerminalRequest, Wait};
use std::{
    io::{self, Write},
    pin::Pin,
    sync::{atomic::{AtomicBool, AtomicUsize, Ordering}, Arc, Mutex},
    task::{Context, Poll},
    time::{Duration, Instant},
};

fn event(text: &str) {
    let path = std::path::PathBuf::from(std::env::var_os("P1_SHUTDOWN_DIRECTORY").unwrap())
        .join("events");
    let mut file = std::fs::OpenOptions::new().create(true).append(true).open(path).unwrap();
    let record = format!("{text}\n");
    file.write_all(record.as_bytes()).unwrap();
}

struct Observer {
    starts: AtomicUsize,
    first: AtomicBool,
    labels: Mutex<std::collections::HashMap<u32, String>>,
    runtime: Mutex<Option<runtime::Runtime>>,
}

impl Dependencies for Observer {
    fn start(&self, command: &mut tokio::process::Command) -> io::Result<tokio::process::Child> {
        let label = if self.starts.fetch_add(1, Ordering::SeqCst) == 0 { "first" } else { "second" };
        command.env("P1_SHUTDOWN_LABEL", label);
        let child = command.spawn()?;
        self.labels.lock().unwrap().insert(child.id().unwrap(), label.into());
        Ok(child)
    }

    fn wait<'a>(&'a self, child: &'a mut tokio::process::Child) -> Wait<'a> {
        let label = self.labels.lock().unwrap().get(&child.id().unwrap()).unwrap().clone();
        Box::pin(async move {
            let result = child.wait().await;
            event(&format!("wait {label}"));
            result
        })
    }

    fn signal(&self, target: SignalTarget, requested: Signal) -> io::Result<()> {
        let mut inject = false;
        if !self.first.swap(true, Ordering::SeqCst) {
            let deadline = Instant::now() + Duration::from_secs(4);
            let dir = std::path::PathBuf::from(std::env::var_os("P1_SHUTDOWN_DIRECTORY").unwrap());
            for label in ["first", "first.descendant", "second", "second.descendant"] {
                while !dir.join(format!("{label}.ready")).exists() {
                    assert!(Instant::now() < deadline, "companion readiness deadline: {label}");
                    std::thread::sleep(Duration::from_millis(1));
                }
            }
            let mode = std::env::var("P1_SHUTDOWN_MODE").unwrap();
            if mode == "sigint" || mode == "sigterm" {
                let (sig, code) = if mode == "sigint" { (Signal::SIGINT, 130) } else { (Signal::SIGTERM, 143) };
                signal::kill(Pid::this(), sig).unwrap();
                while runtime::test_support::terminal_status(self.runtime.lock().unwrap().as_ref().unwrap()) != Some(code as u8) {
                    assert!(Instant::now() < deadline, "signal acceptance deadline");
                    std::thread::yield_now();
                }
                event(&format!("accepted {code}"));
            }
            inject = mode == "cleanup-error";
        }
        let result = match target {
            SignalTarget::Group(id) => signal::killpg(Pid::from_raw(id as i32), requested),
            SignalTarget::Leader(id) => signal::kill(Pid::from_raw(id as i32), requested),
        }.map_err(io::Error::from);
        if inject { return Err(io::Error::other("injected cleanup failure")); }
        result
    }

    fn capture<'a>(&self, _: Stream, destination: Box<dyn io::Write + Send + 'a>) -> Box<dyn io::Write + Send + 'a> {
        destination
    }
    fn checkpoint(&self, _: Checkpoint, _: Arc<runtime::Invocation>) {}
    fn terminal_entered(&self, request: TerminalRequest) {
        if let TerminalRequest::Signal(code) = request { event(&format!("entered {code}")); }
    }
}

struct Output;
impl tokio::io::AsyncWrite for Output {
    fn poll_write(self: Pin<&mut Self>, _: &mut Context<'_>, bytes: &[u8]) -> Poll<io::Result<usize>> {
        let count = std::io::stdout().write(bytes)?;
        std::io::stdout().flush()?;
        match std::env::var("P1_SHUTDOWN_MODE").as_deref() {
            Ok("panic") => panic!("injected source panic"),
            Ok("panic-nil") => std::panic::panic_any(()),
            _ => {}
        }
        Poll::Ready(Ok(count))
    }
    fn poll_flush(self: Pin<&mut Self>, _: &mut Context<'_>) -> Poll<io::Result<()>> {
        Poll::Ready(std::io::stdout().flush())
    }
    fn poll_shutdown(self: Pin<&mut Self>, context: &mut Context<'_>) -> Poll<io::Result<()>> {
        self.poll_flush(context)
    }
}

fn main() -> std::process::ExitCode {
    let executor = tokio::runtime::Builder::new_multi_thread().enable_all().build().unwrap();
    let result = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| executor.block_on(async {
        let observer = Arc::new(Observer {
            starts: AtomicUsize::new(0),
            first: AtomicBool::new(false),
            labels: Mutex::new(std::collections::HashMap::new()),
            runtime: Mutex::new(None),
        });
        let runtime = runtime::test_support::new_observed(observer.clone(), Box::pin(Output), Box::pin(tokio::io::stderr())).unwrap();
        *observer.runtime.lock().unwrap() = Some(runtime.clone());
        generated::l_entry_with_runtime(&runtime).await
    })));
    executor.shutdown_background();
    match result {
        Ok(code) => std::process::ExitCode::from(code),
        Err(payload) => std::panic::resume_unwind(payload),
    }
}
