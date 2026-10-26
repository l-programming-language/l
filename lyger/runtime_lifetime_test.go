package Lyger

import (
	"testing"

	Logiceme "github.com/l-programming-language/l/logiceme"
)

// Build the production runtime values unchanged. Test-owned Rust observes their
// actual Rc targets; it supplies no replacement ownership implementation.
func runtimeLifetimeBundle(t *testing.T) Logiceme.Executable {
	t.Helper()
	null, err := boxRuntimeValue(emittedExpression{kind: valueNull, text: "{ evaluations.set(evaluations.get() + 1); () }"})
	if err != nil {
		t.Fatal(err)
	}
	support := runtimeValueSupport + programRuntimeSupport() + runtimeLifetimeProbe +
		"async fn l_main(runtime: &runtime::Runtime) -> runtime::Result<()> { let evaluations = std::cell::Cell::new(0); let _ = " + null + "; assert_eq!(evaluations.get(), 1); observe_lifetimes(runtime).await }"
	main := newRootObject("main")
	primary := newRootService("main")
	primary.Objects = []*object{main}
	app := &App{ID: "entry", Meta: appMetadata("entry"), Services: []*service{primary}}
	words := []*w{{Syllables: []*s{{CompiledBytes: []byte("A")}, {CompiledBytes: []byte("DA^" + orchestrationEvaluation(support))}}}}
	program := &Program{Books: []*book{{Name: "entry", Chapters: []*chapter{{Name: "main", sentences: []*sentence{{Name: "main", Words: words}}}}}}}
	_, aml, err := renderProgram(program, []*App{app})
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := Logiceme.NewLogicemeService(phase1Options(t).Resources).Evaluate(aml, Logiceme.GenerationOptions{IncludeRuntime: true})
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

func TestRuntimeLifetimeSourceGeneration(t *testing.T) { runtimeLifetimeBundle(t) }

func TestRuntimeLifetimesNative(t *testing.T) {
	runtimeBundleResult(t, runtimeLifetimeBundle(t), "", "")
}

const runtimeLifetimeProbe = `
type Probe = for<'a> fn(&'a (runtime::Runtime, LProgram), LEnvironment) -> std::pin::Pin<Box<dyn std::future::Future<Output = runtime::Result<std::ops::ControlFlow<(), (LValue,)>>> + 'a>>;
fn probe_value(context: &(runtime::Runtime, LProgram), captures: LEnvironment) -> std::pin::Pin<Box<dyn std::future::Future<Output = runtime::Result<std::ops::ControlFlow<(), (LValue,)>>> + '_>> {
    Box::pin(async move {
        let _ = context;
        let value = captures.get(0)?.into_environment()?.get(0)?.into_pointer()?.get()?;
        Ok(std::ops::ControlFlow::Continue((value,)))
    })
}
async fn observe_lifetimes(runtime: &runtime::Runtime) -> runtime::Result<()> {
    use std::rc::Rc;
    // Ordinary Map/Instance snapshots own their fields but share Pointer targets.
    let pointer = LPointer::typed(LType::Thing, LValue::Text(b"retained".to_vec()))?;
    let target = Rc::downgrade(&pointer.target);
    let mut map = LMap::new(LType::Thing);
    map.set(b"key".to_vec(), LValue::Pointer(pointer.clone()))?;
    let mut instance = LInstance::new(1, vec![(b"field".to_vec(), LType::Thing)]);
    instance.set(b"field".to_vec(), LValue::Map(map))?;
    let snapshot = instance.clone();
    drop(pointer);
    drop(instance);
    assert!(target.upgrade().is_some());
    drop(snapshot);
    assert!(target.upgrade().is_none());

    // A callback Map and its copied Function retain nested action environments.
    let pointer = LPointer::typed(LType::Thing, LValue::Text(b"callback".to_vec()))?;
    let target = Rc::downgrade(&pointer.target);
    let action = LEnvironment::typed(vec![LType::Thing], vec![LValue::Pointer(pointer.clone())])?;
    let action_lifetime = Rc::downgrade(&action.values);
    let captures = LEnvironment::typed(vec![LType::Thing], vec![LValue::Environment(action.clone())])?;
    let capture_lifetime = Rc::downgrade(&captures.values);
    let signature = LType::Function(vec![], vec![LType::Thing]);
    let callback = (captures, probe_value as Probe);
    let mut callbacks = LMap::new(signature.clone());
    callbacks.set(b"saved".to_vec(), LValue::function(callback, signature.clone()))?;
    let saved = callbacks.get(b"saved")?.into_function::<(LEnvironment, Probe)>()?;
    let context = (runtime.clone(), LProgram::new(vec![], vec![signature.clone()], 1));
    let boxed = LValue::function(saved.clone(), signature);
    let LValue::Function(ref carrier, _, _) = boxed else { panic!("expected native Function value") };
    let storage_lifetime = Rc::downgrade(carrier);
    context.1.set(0, boxed)?;
    context.1.bind_action(0, action.clone())?;
    drop(callbacks);
    drop(action);
    drop(pointer);
    assert!(target.upgrade().is_some() && action_lifetime.upgrade().is_some());
    assert!(capture_lifetime.upgrade().is_some() && storage_lifetime.upgrade().is_some());
    let std::ops::ControlFlow::Continue(result) = (saved.1)(&context, saved.0.clone()).await? else { panic!("unexpected Exit") };
    assert_eq!(result.0.display_bytes()?, b"callback");
    drop(saved);
    assert!(target.upgrade().is_some() && capture_lifetime.upgrade().is_some());
    drop(context);
    assert!(target.upgrade().is_none() && action_lifetime.upgrade().is_none());
    assert!(capture_lifetime.upgrade().is_none() && storage_lifetime.upgrade().is_none());

    // Rejected writes leave their targets unchanged and introduce no retained cycle.
    let pointer = LPointer::typed(LType::Thing, LValue::Null)?;
    let target = Rc::downgrade(&pointer.target);
    let mut map = LMap::new(LType::Thing);
    map.set(b"edge".to_vec(), LValue::Pointer(pointer.clone()))?;
    assert!(pointer.set(LValue::Map(map)).is_err());
    let mut instance = LInstance::new(2, vec![(b"edge".to_vec(), LType::Thing)]);
    instance.set(b"edge".to_vec(), LValue::Pointer(pointer.clone()))?;
    assert!(pointer.set(LValue::Instance(instance)).is_err());
    assert!(matches!(pointer.get()?, LValue::Null));
    drop(pointer);
    assert!(target.upgrade().is_none());

    let environment = LEnvironment::typed(vec![LType::Thing], vec![LValue::Null])?;
    let lifetime = Rc::downgrade(&environment.values);
    let mut map = LMap::new(LType::Thing);
    map.set(b"edge".to_vec(), LValue::Environment(environment.clone()))?;
    assert!(environment.set(0, LValue::Map(map)).is_err());
    let callback = (environment.clone(), probe_value as Probe);
    assert!(environment.set(0, LValue::function(callback, LType::Function(vec![], vec![LType::Thing]))).is_err());
    assert!(matches!(environment.get(0)?, LValue::Null));
    drop(environment);
    assert!(lifetime.upgrade().is_none());
    Ok(())
}
`
