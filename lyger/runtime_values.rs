// Owned ordinary values snapshot recursively; Pointer and callable environments
// retain identity. All graph mutation goes through checked, short borrows.
#[derive(Clone, PartialEq, Eq)]
enum LType {
    Thing,
    Text,
    Number,
    Truth,
    Bytes,
    Null,
    List(Box<LType>),
    Map(Box<LType>),
    Instance(u64),
    Pointer(Box<LType>),
    Error,
    App,
    Function(Vec<LType>, Vec<LType>),
}
impl LType {
    fn accepts(&self, other: &Self) -> bool {
        matches!(self, Self::Thing) || self == other
    }
}
#[derive(Clone)]
struct LList {
    element: LType,
    values: Vec<LValue>,
}
#[derive(Clone)]
struct LMap {
    element: LType,
    values: std::collections::BTreeMap<Vec<u8>, LValue>,
}
#[derive(Clone)]
struct LInstance {
    identity: u64,
    schema: std::collections::BTreeMap<Vec<u8>, LType>,
    fields: std::collections::BTreeMap<Vec<u8>, LValue>,
}
#[derive(Clone)]
struct LPointer {
    element: LType,
    target: std::rc::Rc<std::cell::RefCell<LValue>>,
}
#[derive(Clone)]
struct LEnvironment {
    values: std::rc::Rc<std::cell::RefCell<Vec<LValue>>>,
    types: std::rc::Rc<Vec<LType>>,
}
#[derive(Clone)]
enum LValue {
    Text(Vec<u8>),
    Number(f64),
    Truth(bool),
    Bytes(Vec<u8>),
    Null,
    List(LList),
    Map(LMap),
    Instance(LInstance),
    Pointer(LPointer),
    Error(runtime::Error),
    App(std::sync::Arc<runtime::App>),
    Function(LCallable),
    Environment(LEnvironment),
}
impl LValue {
    fn expect_type(self, expected: LType) -> runtime::Result<Self> {
        if expected.accepts(&self.ty()) {
            Ok(self)
        } else {
            Err(runtime::Error::new("runtime value type mismatch"))
        }
    }
    fn ty(&self) -> LType {
        match self {
            Self::Text(_) => LType::Text,
            Self::Number(_) => LType::Number,
            Self::Truth(_) => LType::Truth,
            Self::Bytes(_) => LType::Bytes,
            Self::Null => LType::Null,
            Self::List(v) => LType::List(Box::new(v.element.clone())),
            Self::Map(v) => LType::Map(Box::new(v.element.clone())),
            Self::Instance(v) => LType::Instance(v.identity),
            Self::Pointer(v) => LType::Pointer(Box::new(v.element.clone())),
            Self::Error(_) => LType::Error,
            Self::App(_) => LType::App,
            Self::Function(v) => v.signature.clone(),
            Self::Environment(_) => LType::Thing,
        }
    }
    fn reaches(
        &self,
        target: usize,
        seen: &mut std::collections::BTreeSet<usize>,
    ) -> runtime::Result<bool> {
        match self {
            Self::Pointer(v) => {
                let id = std::rc::Rc::as_ptr(&v.target) as usize;
                if id == target {
                    return Ok(true);
                }
                if !seen.insert(id) {
                    return Ok(false);
                }
                v.target
                    .try_borrow()
                    .map_err(|_| runtime::Error::new("Pointer is already borrowed"))?
                    .reaches(target, seen)
            }
            Self::Environment(v) => v.reaches(target, seen),
            Self::Function(v) => v.captures.reaches(target, seen),
            Self::List(v) => {
                for child in &v.values {
                    if child.reaches(target, seen)? {
                        return Ok(true);
                    }
                }
                Ok(false)
            }
            Self::Map(v) => {
                for child in v.values.values() {
                    if child.reaches(target, seen)? {
                        return Ok(true);
                    }
                }
                Ok(false)
            }
            Self::Instance(v) => {
                for child in v.fields.values() {
                    if child.reaches(target, seen)? {
                        return Ok(true);
                    }
                }
                Ok(false)
            }
            _ => Ok(false),
        }
    }
}
macro_rules! l_project {
    ($name:ident, $variant:ident, $ty:ty) => {
        impl LValue {
            fn $name(self) -> runtime::Result<$ty> {
                if let Self::$variant(value) = self {
                    Ok(value)
                } else {
                    Err(runtime::Error::new(concat!(
                        "expected ",
                        stringify!($variant)
                    )))
                }
            }
        }
    };
}
l_project!(into_text, Text, Vec<u8>);
l_project!(into_number, Number, f64);
l_project!(into_truth, Truth, bool);
l_project!(into_bytes, Bytes, Vec<u8>);
l_project!(into_list, List, LList);
l_project!(into_map, Map, LMap);
l_project!(into_instance, Instance, LInstance);
l_project!(into_pointer, Pointer, LPointer);
l_project!(into_error, Error, runtime::Error);
l_project!(into_app, App, std::sync::Arc<runtime::App>);
l_project!(into_function, Function, LCallable);
l_project!(into_environment, Environment, LEnvironment);
impl LValue {
    fn into_null(self) -> runtime::Result<()> {
        if matches!(self, Self::Null) {
            Ok(())
        } else {
            Err(runtime::Error::new("expected Null"))
        }
    }
}
impl LPointer {
    fn typed(element: LType, value: LValue) -> runtime::Result<Self> {
        if !element.accepts(&value.ty()) {
            return Err(runtime::Error::new("Pointer initialization type mismatch"));
        }
        Ok(Self {
            element,
            target: std::rc::Rc::new(std::cell::RefCell::new(value)),
        })
    }
    fn get(&self) -> runtime::Result<LValue> {
        Ok(self
            .target
            .try_borrow()
            .map_err(|_| runtime::Error::new("Pointer is already borrowed"))?
            .clone())
    }
    fn set(&self, value: LValue) -> runtime::Result<()> {
        if !self.element.accepts(&value.ty()) {
            return Err(runtime::Error::new("Pointer target type mismatch"));
        }
        if value.reaches(
            std::rc::Rc::as_ptr(&self.target) as usize,
            &mut std::collections::BTreeSet::new(),
        )? {
            return Err(runtime::Error::new(
                "Pointer write would create an ownership cycle",
            ));
        }
        *self
            .target
            .try_borrow_mut()
            .map_err(|_| runtime::Error::new("Pointer is already borrowed"))? = value;
        Ok(())
    }
}
impl LEnvironment {
    fn typed(types: Vec<LType>, values: Vec<LValue>) -> runtime::Result<Self> {
        if types.len() != values.len()
            || types.iter().zip(&values).any(|(t, v)| !t.accepts(&v.ty()))
        {
            return Err(runtime::Error::new("capture initialization type mismatch"));
        }
        Ok(Self {
            values: std::rc::Rc::new(std::cell::RefCell::new(values)),
            types: std::rc::Rc::new(types),
        })
    }
    fn get(&self, index: usize) -> runtime::Result<LValue> {
        self.values
            .try_borrow()
            .map_err(|_| runtime::Error::new("capture environment is already borrowed"))?
            .get(index)
            .cloned()
            .ok_or_else(|| runtime::Error::new("capture index out of bounds"))
    }
    fn set(&self, index: usize, value: LValue) -> runtime::Result<()> {
        if value.reaches(
            std::rc::Rc::as_ptr(&self.values) as usize,
            &mut std::collections::BTreeSet::new(),
        )? {
            return Err(runtime::Error::new(
                "capture write would create an ownership cycle",
            ));
        }
        let mut values = self
            .values
            .try_borrow_mut()
            .map_err(|_| runtime::Error::new("capture environment is already borrowed"))?;
        let current = values
            .get_mut(index)
            .ok_or_else(|| runtime::Error::new("capture index out of bounds"))?;
        if !self.types[index].accepts(&value.ty()) {
            return Err(runtime::Error::new("capture type mismatch"));
        }
        *current = value;
        Ok(())
    }
    fn reaches(
        &self,
        target: usize,
        seen: &mut std::collections::BTreeSet<usize>,
    ) -> runtime::Result<bool> {
        let id = std::rc::Rc::as_ptr(&self.values) as usize;
        if id == target {
            return Ok(true);
        }
        if !seen.insert(id) {
            return Ok(false);
        }
        for value in self
            .values
            .try_borrow()
            .map_err(|_| runtime::Error::new("capture environment is already borrowed"))?
            .iter()
        {
            if value.reaches(target, seen)? {
                return Ok(true);
            }
        }
        Ok(false)
    }
}
fn l_index(index: f64) -> runtime::Result<usize> {
    if !index.is_finite() || index < 0.0 || index.fract() != 0.0 || index >= usize::MAX as f64 {
        Err(runtime::Error::new("index must be a nonnegative integer"))
    } else {
        Ok(index as usize)
    }
}
impl LList {
    fn new(element: LType, values: Vec<LValue>) -> runtime::Result<Self> {
        if values.iter().any(|v| !element.accepts(&v.ty())) {
            return Err(runtime::Error::new("List element type mismatch"));
        }
        Ok(Self { element, values })
    }
    fn get(&self, index: f64) -> runtime::Result<LValue> {
        self.values
            .get(l_index(index)?)
            .cloned()
            .ok_or_else(|| runtime::Error::new("List index out of bounds"))
    }
    fn put(&mut self, value: LValue) -> runtime::Result<()> {
        if !self.element.accepts(&value.ty()) {
            return Err(runtime::Error::new("List element type mismatch"));
        }
        self.values.push(value);
        Ok(())
    }
    fn set(&mut self, index: f64, value: LValue) -> runtime::Result<()> {
        if !self.element.accepts(&value.ty()) {
            return Err(runtime::Error::new("List element type mismatch"));
        }
        *self
            .values
            .get_mut(l_index(index)?)
            .ok_or_else(|| runtime::Error::new("List index out of bounds"))? = value;
        Ok(())
    }
}
impl LMap {
    fn new(element: LType) -> Self {
        Self {
            element,
            values: std::collections::BTreeMap::new(),
        }
    }
    fn get(&self, key: &[u8]) -> runtime::Result<LValue> {
        self.values
            .get(key)
            .cloned()
            .ok_or_else(|| runtime::Error::new("Map key is absent"))
    }
    fn set(&mut self, key: Vec<u8>, value: LValue) -> runtime::Result<()> {
        if !self.element.accepts(&value.ty()) {
            return Err(runtime::Error::new("Map element type mismatch"));
        }
        self.values.insert(key, value);
        Ok(())
    }
}
impl LInstance {
    fn new(identity: u64, fields: Vec<(Vec<u8>, LType)>) -> Self {
        Self {
            identity,
            schema: fields.into_iter().collect(),
            fields: std::collections::BTreeMap::new(),
        }
    }
    fn get(&self, key: &[u8]) -> runtime::Result<LValue> {
        self.fields
            .get(key)
            .cloned()
            .ok_or_else(|| runtime::Error::new("Instance field is absent"))
    }
    fn set(&mut self, key: Vec<u8>, value: LValue) -> runtime::Result<()> {
        if !self
            .schema
            .get(&key)
            .ok_or_else(|| runtime::Error::new("unknown Instance field"))?
            .accepts(&value.ty())
        {
            return Err(runtime::Error::new("Instance field type mismatch"));
        }
        self.fields.insert(key, value);
        Ok(())
    }
}
impl LValue {
    fn display_bytes(&self) -> runtime::Result<Vec<u8>> {
        Ok(match self {
            Self::Text(v) | Self::Bytes(v) => v.clone(),
            Self::Number(v) => runtime::format_number(*v),
            Self::Truth(v) => v.to_string().into_bytes(),
            Self::Null => b"null".to_vec(),
            Self::Error(v) => v.to_string().into_bytes(),
            Self::App(_) => b"App".to_vec(),
            Self::Function(_) => b"Function".to_vec(),
            Self::Environment(_) => {
                return Err(runtime::Error::new("capture environment is not an L value"))
            }
            Self::Pointer(v) => v.get()?.display_bytes()?,
            Self::List(v) => {
                let mut out = b"[".to_vec();
                for (i, child) in v.values.iter().enumerate() {
                    if i > 0 {
                        out.extend_from_slice(b", ");
                    }
                    out.extend(child.display_bytes()?);
                }
                out.extend_from_slice(b"]");
                out
            }
            Self::Map(v) => l_fields(&v.values)?,
            Self::Instance(v) => l_fields(&v.fields)?,
        })
    }
}
fn l_fields(fields: &std::collections::BTreeMap<Vec<u8>, LValue>) -> runtime::Result<Vec<u8>> {
    let mut out = b"{".to_vec();
    for (i, (key, value)) in fields.iter().enumerate() {
        if i > 0 {
            out.extend_from_slice(b", ");
        }
        out.extend(key);
        out.extend_from_slice(b": ");
        out.extend(value.display_bytes()?);
    }
    out.extend_from_slice(b"}");
    Ok(out)
}
