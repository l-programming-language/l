// Generated as process_other.rs. No process-group guarantees are claimed here.
use std::io;
use std::path::Path;
use std::process::ExitStatus;
use tokio::process::Command;
use tokio::time::Instant;

fn unsupported() -> io::Error {
    io::Error::new(io::ErrorKind::Unsupported, "L native applications require macOS or Linux")
}

pub(crate) struct Signals;

impl Signals {
    pub(crate) fn new() -> io::Result<Self> { Err(unsupported()) }
    pub(crate) async fn recv(&mut self) -> io::Result<u8> { Err(unsupported()) }
}

pub(crate) fn configure(_: &mut Command) -> io::Result<()> { Err(unsupported()) }
pub(crate) fn validate_executable(_: &Path) -> io::Result<()> { Err(unsupported()) }
pub(crate) fn terminate_group(_: u32) -> io::Result<()> { Err(unsupported()) }
pub(crate) async fn final_group_result(
    _: u32,
    _: io::Result<()>,
    _: Option<Instant>,
) -> io::Result<()> { Err(unsupported()) }
pub(crate) fn status_code(status: ExitStatus) -> Option<i32> { status.code() }
