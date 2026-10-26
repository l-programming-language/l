package Lyger

func functionRuntimeSupport() string {
	return `
// Bare Exit reaches the program entry without claiming terminal status before cleanup.
enum LFunctionOutcome { Returned(Vec<LValue>), Exit }

#[derive(Clone)]
struct LCallable {
    signature: LType,
    captures: LEnvironment,
    invoke: std::rc::Rc<dyn for<'a> Fn(&'a runtime::Runtime, &'a LProgram, LCallable, Vec<LValue>) -> std::pin::Pin<Box<dyn std::future::Future<Output = runtime::Result<LFunctionOutcome>> + 'a>>>,
}
impl LCallable {
    async fn call(&self, runtime: &runtime::Runtime, program: &LProgram, arguments: Vec<LValue>) -> runtime::Result<LFunctionOutcome> {
        let LType::Function(inputs, outputs) = &self.signature else { return Err(runtime::Error::new("invalid Function signature")); };
        if inputs.len() != arguments.len() || inputs.iter().zip(&arguments).any(|(ty, value)| !ty.accepts(&value.ty())) {
            return Err(runtime::Error::new("Function argument type mismatch"));
        }
        let LFunctionOutcome::Returned(values) = (self.invoke)(runtime, program, self.clone(), arguments).await? else {
            return Ok(LFunctionOutcome::Exit);
        };
        if outputs.len() != values.len() || outputs.iter().zip(&values).any(|(ty, value)| !ty.accepts(&value.ty())) {
            return Err(runtime::Error::new("Function result type mismatch"));
        }
        Ok(LFunctionOutcome::Returned(values))
    }
}
`
}
