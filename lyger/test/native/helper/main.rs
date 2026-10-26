#[path = "ownership.rs"] mod ownership;
use std::io::Write;
fn main() {
    if std::env::var_os("P1_SHUTDOWN_DIRECTORY").is_none() {
        let path=std::env::current_exe().unwrap().parent().unwrap().join("seen");
        if !path.exists() {std::fs::write(path,b"seen").unwrap();std::io::stdout().write_all(b"first\n").unwrap();}
        else {std::io::stdout().write_all(b"second\n").unwrap();}
        return;
    }
    let dir=std::path::PathBuf::from(std::env::var_os("P1_SHUTDOWN_DIRECTORY").unwrap());
    let mut label=std::env::var("P1_SHUTDOWN_LABEL").unwrap();
    let descendant=std::env::args().len()>1;
    if descendant {label.push_str(".descendant");}
    let Some((_held,lifetime))=ownership::register(&label).unwrap() else {return};
    std::io::stdout().write_all(b"child stdout\0\n").unwrap();
    std::io::stderr().write_all(b"child stderr\0\n").unwrap();
    let ready=dir.join(format!("{label}.ready"));let temporary=ready.with_extension("ready.tmp");
    std::fs::write(&temporary,format!("{} {} {}",lifetime.pid,lifetime.pgid,lifetime.record)).unwrap();
    std::fs::rename(temporary,ready).unwrap();
    if !descendant {std::process::Command::new(std::env::current_exe().unwrap()).arg("descendant").spawn().unwrap();}
    loop {std::thread::sleep(std::time::Duration::from_secs(3600));}
}
