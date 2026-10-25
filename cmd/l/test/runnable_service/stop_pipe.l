Define "Log|-s|-ged|-ging|-ger" to `<DEF>BOD^E{|Statement: ~{
    use ::std::io::{BufRead, Write};
    use ::std::os::unix::fs::OpenOptionsExt;
    let value = <@Output|Thing@>;
    if value == \"worker\" {
        let status = ::std::process::Command::new(::std::env::var(\"L_SERVICE_PIPE_HELPER\").unwrap())
            .arg(\"-test.run=^TestRunnableServicePipeBridgeHelper$\")
            .env(\"L_SERVICE_PIPE_BRIDGE_CHILD\", \"1\")
            .stdin(::std::process::Stdio::null())
            .stdout(::std::process::Stdio::inherit())
            .stderr(::std::process::Stdio::inherit())
            .status().unwrap();
        assert!(status.success(), \"Service pipe transfer failed\");
    }
    let lifetime = ::std::fs::OpenOptions::new().create(true).read(true).write(true).mode(0o600)
        .open(::std::path::Path::new(&::std::env::var(\"L_PHASE1_DIRECTORY\").unwrap()).join(format!(\"alive-{}-{}\", value, ::std::process::id()))).unwrap();
    lifetime.try_lock().unwrap();
    let mut connection = ::std::os::unix::net::UnixStream::connect(::std::env::var(\"L_PHASE1_SOCKET\").unwrap()).unwrap();
    writeln!(connection, \"ready {} {}\", value, ::std::process::id()).unwrap();
    let mut command = String::new();
    ::std::io::BufReader::new(&connection).read_line(&mut command).unwrap();
    assert_eq!(command, \"release\n\", \"unexpected controller command\");
    assert_eq!(value, \"gate\", \"Worker must be stopped without release\");
    drop(connection);
    drop(lifetime);
}~|}<DEF>`.
Create "StopProbe" as a Dictionary.
Create StopProbe's "Print|-s|-ed|-ing|-er" for Processing a New Logger.
