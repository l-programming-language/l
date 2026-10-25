use std::collections::BTreeMap;
use std::ffi::OsStr;
use std::io;

// Key and value inputs are complete, escaped Rust string literals. The map
// resolves repeated keys before comparing the inherited process environment.
pub fn initialize_configuration_settings() -> io::Result<()> {
    let mut settings: BTreeMap<&str, &str> = BTreeMap::new();
    
    settings.insert("Host", "http://127.0.0.1");
    

    for (key, value) in &settings {
        if key.is_empty() || key.contains('=') || key.contains('\0') || value.contains('\0') {
            return Err(io::Error::new(
                io::ErrorKind::InvalidInput,
                "invalid configuration environment key or value",
            ));
        }
    }

    #[cfg(any(target_os = "linux", target_os = "macos"))]
    {
        use std::os::unix::process::CommandExt;

        if settings.iter().all(|(key, value)| {
            std::env::var_os(key).as_deref() == Some(OsStr::new(value))
        }) {
            return Ok(());
        }

        // This function runs before Tokio, signal handlers, capture tasks, or
        // application resources exist. exec keeps this PID, CWD and streams.
        let executable = std::env::current_exe()?;
        let mut arguments = std::env::args_os();
        let mut command = std::process::Command::new(executable);
        if let Some(argument_zero) = arguments.next() {
            command.arg0(argument_zero);
        }
        command.args(arguments).envs(settings);
        let error = command.exec();
        Err(io::Error::new(
            error.kind(),
            format!("configuration startup re-execution failed: {error}"),
        ))
    }

    #[cfg(not(any(target_os = "linux", target_os = "macos")))]
    {
        Err(io::Error::new(
            io::ErrorKind::Unsupported,
            "generated L applications support native macOS and Linux only",
        ))
    }
}
