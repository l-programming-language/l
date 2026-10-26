Define "Log|-s|-ged|-ging|-ger" to `<DEF>BOD^E{|Statement: ~{
    use ::std::io::Write;
    ::std::io::stdout().write_all(<@Output|Thing@>.as_bytes()).unwrap();
    ::std::io::stderr().write_all(b\"problem\n\").unwrap();
    ::std::io::stdout().flush().unwrap();
    ::std::io::stderr().flush().unwrap();
    ::std::process::exit(7);
}~|}<DEF>`.
Use `actions.l`.
