Define "Log|-s|-ged|-ging|-ger" to `<DEF>BOD^E{|Statement: ~{
    use ::std::io::Write;
    use ::std::os::unix::fs::OpenOptionsExt;
    let path = ::std::path::Path::new(&::std::env::var(\"L_PHASE1_DIRECTORY\").unwrap()).join(\"count\");
    let count: i64 = match ::std::fs::read_to_string(&path) {
        Ok(raw) => raw.parse::<i64>().unwrap(),
        Err(error) if error.kind() == ::std::io::ErrorKind::NotFound => 0,
        Err(error) => panic!(\"{}\", error),
    } + 1;
    let mut file = ::std::fs::OpenOptions::new().create(true).write(true).truncate(true).mode(0o600).open(path).unwrap();
    write!(file, \"{}\", count).unwrap();
    ::std::println!(\"{}#{}\", <@Output|Thing@>, count);
}~|}<DEF>`.
Use `actions.l`.
