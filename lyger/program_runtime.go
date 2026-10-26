package Lyger

func programRuntimeSupport() string {
	return `
// Program state is borrowed by a call, never captured or boxed as an L value.
// None tracks uninitialized compiler storage; it is not an L Null value.
struct LProgram {
    apps: Vec<std::sync::Arc<runtime::App>>,
    types: Vec<LType>,
    values: std::cell::RefCell<Vec<Option<LValue>>>,
    actions: std::cell::RefCell<Vec<Option<LEnvironment>>>,
}
impl LProgram {
    fn new(apps: Vec<std::sync::Arc<runtime::App>>, types: Vec<LType>, action_count: usize) -> Self {
        let values = std::cell::RefCell::new(vec![None; types.len()]);
        Self { apps, types, values, actions: std::cell::RefCell::new(vec![None; action_count]) }
    }
    fn get(&self, index: usize) -> runtime::Result<LValue> {
        self.values.try_borrow().map_err(|_| runtime::Error::new("program storage is already borrowed"))?
            .get(index).and_then(Option::as_ref).cloned().ok_or_else(|| runtime::Error::new("program value is not initialized"))
    }
    fn set(&self, index: usize, value: LValue) -> runtime::Result<()> {
        let ty = self.types.get(index).ok_or_else(|| runtime::Error::new("program storage index out of bounds"))?;
        if !ty.accepts(&value.ty()) { return Err(runtime::Error::new("program value type mismatch")); }
        self.values.try_borrow_mut().map_err(|_| runtime::Error::new("program storage is already borrowed"))?[index] = Some(value);
        Ok(())
    }
    fn action(&self, index: usize) -> runtime::Result<LEnvironment> {
        self.actions.try_borrow().map_err(|_| runtime::Error::new("program actions are already borrowed"))?
            .get(index).and_then(Option::as_ref).cloned().ok_or_else(|| runtime::Error::new("program action is not initialized"))
    }
    fn bind_action(&self, index: usize, value: LEnvironment) -> runtime::Result<()> {
        let mut actions = self.actions.try_borrow_mut().map_err(|_| runtime::Error::new("program actions are already borrowed"))?;
        let target = actions.get_mut(index).ok_or_else(|| runtime::Error::new("program action index out of bounds"))?;
        *target = Some(value);
        Ok(())
    }
}
`
}
