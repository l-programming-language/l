Define "Log|-s|-ged|-ging|-ger" to `<DEF>BOD^E{|Statement: ~{
    use ::std::io::{BufRead, Write};
    use ::std::os::unix::fs::OpenOptionsExt;
    let value = <@Output|Thing@>;
    let lifetime = ::std::fs::OpenOptions::new().create(true).read(true).write(true).mode(0o600)
        .open(::std::path::Path::new(&::std::env::var(\"L_PHASE1_DIRECTORY\").unwrap()).join(format!(\"alive-{}-{}\", value, ::std::process::id()))).unwrap();
    lifetime.try_lock().unwrap();
    if value != \"gate\" {
        ::std::io::stdout().write_all(value.repeat(262144).as_bytes()).unwrap();
        ::std::io::stdout().flush().unwrap();
        ::std::io::stderr().write_all(\"e\".repeat(262144).as_bytes()).unwrap();
    }
    let mut connection = ::std::os::unix::net::UnixStream::connect(::std::env::var(\"L_PHASE1_SOCKET\").unwrap()).unwrap();
    writeln!(connection, \"ready {} {}\", value, ::std::process::id()).unwrap();
    let mut command = String::new();
    ::std::io::BufReader::new(&connection).read_line(&mut command).unwrap();
    assert_eq!(command, \"release\n\", \"unexpected controller command\");
    drop(connection);
    drop(lifetime);
}~|}<DEF>`.
Use `actions.l`.
