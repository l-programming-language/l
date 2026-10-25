Define "Log|-s|-ged|-ging|-ger" to `<DEF>BOD^E{|Statement: ~{
    use ::std::io::Write;
    ::std::io::stdout().write_all(<@Output|Thing@>.as_bytes()).unwrap();
    ::std::io::stdout().write_all(b\" \x00\n\n\").unwrap();
    ::std::io::stdout().flush().unwrap();
    ::std::io::stderr().write_all(b\"diagnostic only\n\").unwrap();
}~|}<DEF>`.
Use `actions.l`.
