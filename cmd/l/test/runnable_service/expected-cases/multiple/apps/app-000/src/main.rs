#![forbid(unsafe_code)]
mod config;


// package main




fn main( ) {
    
        

if let Err(error) = config::initialize_configuration_settings() { ::std::eprintln!("configuration startup: {error}"); ::std::process::exit(1); }



    
        

std::io::Write::write_all(&mut std::io::stdout(), ("first\n").as_bytes()).unwrap();



    
}
