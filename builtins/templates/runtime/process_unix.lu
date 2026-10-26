// Generated as process_unix.rs; selected only on macOS and Linux.
use std::io::{self, Read};
use std::os::unix::fs::PermissionsExt;
use std::path::Path;
use std::process::ExitStatus;

use nix::errno::Errno;
use nix::sys::signal::{kill, killpg, Signal};
use nix::unistd::Pid;
use tokio::process::Command;
use tokio::signal::unix::{signal, Signal as SignalStream, SignalKind};
use tokio::time::{sleep, Duration, Instant};

pub(crate) struct Signals {
    interrupt: SignalStream,
    terminate: SignalStream,
    pipe: SignalStream,
}

impl Signals {
    pub(crate) fn new() -> io::Result<Self> {
        Ok(Self {
            interrupt: signal(SignalKind::interrupt())?,
            terminate: signal(SignalKind::terminate())?,
            pipe: signal(SignalKind::pipe())?,
        })
    }

    pub(crate) async fn recv(&mut self) -> io::Result<u8> {
        let (event, status) = tokio::select! {
            event = self.interrupt.recv() => (event, 130),
            event = self.terminate.recv() => (event, 143),
            event = self.pipe.recv() => (event, 1),
        };
        event.map(|()| status).ok_or_else(|| {
            io::Error::new(io::ErrorKind::UnexpectedEof, "application signal subscription closed")
        })
    }
}

pub(crate) fn configure(command: &mut Command) -> io::Result<()> {
    command.process_group(0);
    Ok(())
}

pub(crate) fn validate_executable(path: &Path) -> io::Result<()> {
    let metadata = std::fs::symlink_metadata(path)?;
    if !metadata.file_type().is_file() || metadata.permissions().mode() & 0o111 == 0 {
        return Err(io::Error::new(io::ErrorKind::PermissionDenied, "not a regular executable"));
    }
    Ok(())
}

// std::process may ask a shell to interpret ENOEXEC inputs on Unix. L Run
// accepts native images or an explicit shebang, never implicit shell programs.
// Build remains a file-readiness check; this check belongs to each launch.
pub(crate) fn validate_launch_format(path: &Path) -> io::Result<()> {
    let mut prefix = [0u8; 4];
    let mut file = std::fs::File::open(path)?;
    let count = file.read(&mut prefix)?;
    if count >= 2 && &prefix[..2] == b"#!" { return Ok(()); }
    let native = if cfg!(target_os = "linux") {
        prefix == *b"\x7fELF"
    } else {
        matches!(u32::from_be_bytes(prefix), 0xfeedface | 0xcefaedfe | 0xfeedfacf | 0xcffaedfe | 0xcafebabe | 0xbebafeca | 0xcafebabf | 0xbfbafeca)
    };
    if count == 4 && native { Ok(()) } else { Err(io::Error::from_raw_os_error(Errno::ENOEXEC as i32)) }
}

fn group(pid: u32) -> io::Result<Pid> {
    let pid = i32::try_from(pid)
        .ok()
        .filter(|pid| *pid > 1)
        .ok_or_else(|| io::Error::new(io::ErrorKind::InvalidInput, "invalid child process group"))?;
    Ok(Pid::from_raw(pid))
}

pub(crate) fn terminate_group(pid: u32) -> io::Result<()> {
    match killpg(group(pid)?, Signal::SIGKILL) {
        Ok(()) | Err(Errno::ESRCH) => Ok(()),
        Err(error) => Err(io::Error::from_raw_os_error(error as i32)),
    }
}

pub(crate) fn terminate_leader(pid: u32) -> io::Result<()> {
    match kill(group(pid)?, Signal::SIGKILL) {
        Ok(()) | Err(Errno::ESRCH) => Ok(()),
        Err(error) => Err(io::Error::from_raw_os_error(error as i32)),
    }
}

// Only a prior successful Stop/shutdown request supplies this deadline. The
// Darwin probe never repeats SIGKILL, renews the deadline, or treats EPERM as
// proof that the group disappeared. The supervisor remains the sole Wait owner.
pub(crate) async fn final_group_result(
    pid: u32,
    result: io::Result<()>,
    deadline: Option<Instant>,
) -> io::Result<()> {
    let error = match result {
        Ok(()) => return Ok(()),
        Err(error) => error,
    };
    if !cfg!(target_os = "macos") || error.raw_os_error() != Some(Errno::EPERM as i32) {
        return Err(error);
    }
    let Some(deadline) = deadline else { return Err(error) };
    let group = group(pid)?;
    while Instant::now() < deadline {
        let observed = killpg(group, None);
        if Instant::now() >= deadline {
            break;
        }
        match observed {
            Err(Errno::ESRCH) => return Ok(()),
            Err(Errno::EPERM) => {}
            _ => return Err(error),
        }
        sleep(deadline.saturating_duration_since(Instant::now()).min(Duration::from_millis(1))).await;
    }
    Err(error)
}

pub(crate) fn status_code(status: ExitStatus) -> Option<i32> {
    use std::os::unix::process::ExitStatusExt;
    status.code().or_else(|| status.signal().map(|signal| 128 + signal))
}
