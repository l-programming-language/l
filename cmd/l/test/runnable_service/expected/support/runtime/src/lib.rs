#![forbid(unsafe_code)]

pub mod runtime;
#[cfg(any(target_os = "linux", target_os = "macos"))]
#[path = "process_unix.rs"]
mod process;
#[cfg(not(any(target_os = "linux", target_os = "macos")))]
#[path = "process_other.rs"]
mod process;
