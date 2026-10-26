#![forbid(unsafe_code)]
mod config;


// package main


use ::l_runtime_support::runtime;
use std::io::Write;
use std::panic::{catch_unwind, resume_unwind, AssertUnwindSafe};
use std::process::ExitCode;


// The source assembler supplies a same-module generated item with this contract:
// async fn l_main(runtime: &runtime::Runtime) -> runtime::Result<()>.
// It does not supply another main function or install another Tokio runtime.
fn main() -> ExitCode {
    if let Err(error) = config::initialize_configuration_settings() {
        let _ = writeln!(std::io::stderr().lock(), "l: {error}");
        return ExitCode::FAILURE;
    }

    let executor = match tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()
    {
        Ok(executor) => executor,
        Err(error) => {
            let _ = writeln!(std::io::stderr().lock(), "l: runtime startup failed: {error}");
            return ExitCode::FAILURE;
        }
    };

    let result = catch_unwind(AssertUnwindSafe(|| executor.block_on(async {
        match runtime::Runtime::new() {
            Ok(runtime) => runtime.run(l_main(&runtime)).await,
            Err(error) => {
                let _ = writeln!(std::io::stderr().lock(), "l: {error}");
                1
            }
        }
    })));

    // Runtime::run settles owned children and captures before returning or
    // resuming a source panic. A cancelled standard-stream write may still be
    // blocked in Tokio's blocking pool; executor Drop must not delay process
    // exit waiting for that unrelated work. No child cleanup is delegated here.
    executor.shutdown_background();
    match result {
        Ok(status) => ExitCode::from(status),
        Err(panic) => resume_unwind(panic),
    }
}

fn l_main( runtime: &runtime::Runtime) -> (impl std::future::Future<Output = runtime::Result<()>> + '_) {
    
        


async move {

let lApp0 = runtime.app("main.apps/generations/@@GENERATION@@/app-000");



lApp0.set_source("@@SOURCE@@:9:1: run Worker");




lApp0.run()?;




lApp0.set_source("@@SOURCE@@:10:1: await Worker");



let lInvocation0 = lApp0.wait().await?;





if lInvocation0.claim_continuation()? {
    

{






if true {
    

return Ok(());



} 






runtime.show(&"unreachable".as_bytes().to_vec()).await?;




}




lInvocation0.finish_continuation();



} 






runtime.show(&"after".as_bytes().to_vec()).await?;




Ok(())



}




    
}


#[cfg(test)]
pub(crate) async fn l_entry_with_runtime(runtime: &runtime::Runtime) -> u8 {
    runtime.run(l_main(runtime)).await
}

