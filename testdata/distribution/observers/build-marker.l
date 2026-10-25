Define "RecordBuild|-s|-ed|-ing|-er" to `<DEF>BOD^E{|Statement: ~{ use ::std::io::Write as _; use ::std::os::unix::fs::OpenOptionsExt as _; let mut file = ::std::fs::OpenOptions::new().create(true).truncate(true).write(true).mode(0o600).open(::std::env::var(\"L_BUILD_SIDE_EFFECT\").unwrap()).unwrap(); file.write_all(format!(\"{}\", <@Output|Thing@>).as_bytes()).unwrap(); drop(file); };~|}<DEF>`.
Create "BuildObserver" as a Dictionary.
Create BuildObserver's "Print|-s|-ed|-ing|-er" for Processing a New RecordBuilder.
