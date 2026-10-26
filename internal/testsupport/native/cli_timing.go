package native

// This code is compiled into the existing standalone provenance recorder. It
// runs real commands only; it cannot produce a successful Cargo result or
// substitute an executable. Keeping the driver in one process avoids a timing
// subprocess before every phase and retains the CLI's process-group ownership.
// The command driver is also used without timing; instrumentation stays opt-in.
const artifactTimingSource = `
var timingSequence uint64
func timingWrite(event map[string]any) {
 dir:=os.Getenv("L_HARNESS_TIMING_DIR");if dir==""{return}
 event["pid"]=os.Getpid();event["ppid"]=os.Getppid();event["run_id"]=proxyEvidenceRunID;if event["run_id"]==""{event["run_id"]=os.Getenv("L_HARNESS_RUN_ID")}
 data,err:=json.Marshal(event);if err!=nil{return}
 if os.MkdirAll(dir,0700)!=nil{return}
 f,err:=os.OpenFile(filepath.Join(dir,fmt.Sprintf("%d.jsonl",os.Getpid())),os.O_CREATE|os.O_WRONLY|os.O_APPEND,0600);if err!=nil{return}
 _,_=f.Write(append(data,'\n'));_=f.Close()
}
func timingStart(phase,subject string) func(error){
 if os.Getenv("L_HARNESS_TIMING_DIR")==""{return func(error){}}
 started:=time.Now();timingSequence++;span:=fmt.Sprintf("%d-recorder-%d",os.Getpid(),timingSequence)
 event:=map[string]any{"event":"start","span":span,"phase":phase,"subject":subject,"started_at":started.UTC(),"duration_ns":0,"duration_seconds":0,"outcome":"running"}
 timingWrite(event);done:=false
 return func(err error){if done{return};done=true;elapsed:=time.Since(started);event["event"]="finish";event["duration_ns"]=elapsed.Nanoseconds();event["duration_seconds"]=elapsed.Seconds();event["outcome"]="ok";if err!=nil{event["outcome"]="error";event["error"]=err.Error()};timingWrite(event)}
}
func timingStats(base,target string,started time.Time){
 if os.Getenv("L_HARNESS_TIMING_DIR")==""{return}
 f,err:=os.Open(base+".jsonl");if err!=nil{return};defer f.Close()
 fresh,rebuilt:=0,0;scan:=bufio.NewScanner(f);scan.Buffer(make([]byte,4096),8*1024*1024)
 for scan.Scan(){var m struct{Reason string;Fresh bool};if json.Unmarshal(scan.Bytes(),&m)==nil&&m.Reason=="compiler-artifact"{if m.Fresh{fresh++}else{rebuilt++}}}
 fields:=map[string]any{"fresh":fresh,"rebuilt":rebuilt}
 if target!=""{source:=filepath.Join(target,"cargo-timings","cargo-timing.html");if info,err:=os.Stat(source);err==nil&&!info.ModTime().Before(started){if data,err:=os.ReadFile(source);err==nil{dir:=filepath.Join(os.Getenv("L_HARNESS_TIMING_DIR"),"cargo-timings");if os.MkdirAll(dir,0700)==nil{dest:=filepath.Join(dir,fmt.Sprintf("%d-%s.html",os.Getpid(),filepath.Base(base)));if os.WriteFile(dest,data,0600)==nil{fields["report"]=dest}}}}}
 timingWrite(map[string]any{"event":"record","phase":"cargo.artifacts","subject":base,"started_at":time.Now().UTC(),"duration_ns":0,"duration_seconds":0,"outcome":"ok","fields":fields})
}
func commandStatus(err error) int{if err==nil{return 0};if exit,ok:=err.(*exec.ExitError);ok{if code:=exit.ExitCode();code>=0{return code};if status,ok:=exit.Sys().(syscall.WaitStatus);ok&&status.Signaled(){return 128+int(status.Signal())}};if os.IsNotExist(err){return 127};if os.IsPermission(err){return 126};return 96}
func observerFailure(cargoErr,observerErr error) int{fmt.Fprintln(os.Stderr,observerErr);if cargoErr!=nil{return commandStatus(cargoErr)};return 96}
func timedCargo(args []string) int{
 if len(args)<2{return 96};base,cargo:=args[0],args[1];native:=append([]string(nil),args[2:]...)
 manifestArgs:=[]string{};target:=""
 for i,arg:=range native{if arg=="--manifest-path"&&i+1<len(native){manifestArgs=[]string{arg,native[i+1]}};if strings.HasPrefix(arg,"--manifest-path="){manifestArgs=[]string{arg}};if arg=="--target-dir"&&i+1<len(native){target=native[i+1]};if strings.HasPrefix(arg,"--target-dir="){target=strings.TrimPrefix(arg,"--target-dir=")}}
 if os.Getenv("L_HARNESS_TIMING_DIR")!=""{native=append(native,"--timings")}
 stream,err:=os.Create(base+".jsonl");if err!=nil{fmt.Fprintln(os.Stderr,err);return 96}
 started:=time.Now();var finish func(error)
 cmd:=exec.Command(cargo,native...);cmd.Stdin=os.Stdin;cmd.Stdout=io.MultiWriter(os.Stdout,stream);cmd.Stderr=os.Stderr
 manifest:="";if len(manifestArgs)==2{manifest=manifestArgs[1]}else if len(manifestArgs)==1{manifest=strings.TrimPrefix(manifestArgs[0],"--manifest-path=")};if manifest!=""{manifest,_=filepath.Abs(manifest)};cwd,_:=os.Getwd()
 workspace:=filepath.Dir(manifest);var located map[string]any;var locateErr error;if os.Getenv("L_HARNESS_CARGO_EVIDENCE_DIR")!=""{workspace,located,locateErr=evidenceLocate(cargo,manifest,cwd)}
 finishEvidence:=timingStart("evidence.inputs",base);capture:=evidenceBegin(evidenceInput{Route:"cli-proxy",Directory:cwd,Workspace:workspace,Manifest:manifest,Target:target,RunID:proxyEvidenceRunID,Test:proxyEvidenceTest,OriginalWorkspace:cwd,SelectedSlot:filepath.Base(filepath.Dir(target)),Toolchain:"1.98.1",Host:proxyEvidenceHost,Command:cmd.Args,Environment:os.Environ(),PID:os.Getpid(),PPID:os.Getppid()});finishEvidence(nil)
 if capture!=nil{if locateErr!=nil{capture.State.Errors=append(capture.State.Errors,locateErr.Error())};if capture.Directory!=""{if err:=evidenceJSON(filepath.Join(capture.Directory,"locate-project.json"),located);err!=nil{capture.State.Errors=append(capture.State.Errors,err.Error())}}}
 if capture!=nil{if writer:=capture.stderr();writer!=nil{cmd.Stderr=io.MultiWriter(os.Stderr,writer)}}
 finish=timingStart("cargo.actual",base)
 cargoErr:=cmd.Run();finish(cargoErr);closeErr:=stream.Close()
 if capture!=nil{defer func(){done:=timingStart("evidence.complete",base);nativeError:="";if cargoErr!=nil{nativeError=cargoErr.Error()};capture.finish(nil,nil,commandStatus(cargoErr),nativeError,base);done(nil)}()}
 if cargoErr!=nil{if _,exited:=cargoErr.(*exec.ExitError);!exited{fmt.Fprintln(os.Stderr,cargoErr)}}
 metadata,err:=os.Create(base+".metadata.json");if err!=nil{return observerFailure(cargoErr,err)}
 diag,err:=os.Create(base+".metadata.stderr");if err!=nil{_=metadata.Close();return observerFailure(cargoErr,err)}
 metadataArgs:=append([]string{"metadata"},manifestArgs...);metadataArgs=append(metadataArgs,"--frozen","--offline","--no-deps","--format-version","1")
 finish=timingStart("cargo.metadata",base)
 cmd=exec.Command(cargo,metadataArgs...);cmd.Stdin=os.Stdin;cmd.Stdout=metadata;cmd.Stderr=diag;err=cmd.Run();finish(err);metadataClose,diagClose:=metadata.Close(),diag.Close();if capture!=nil{for _,captureErr:=range []error{err,metadataClose,diagClose}{if captureErr!=nil{capture.State.Errors=append(capture.State.Errors,captureErr.Error())}}}
 finish=timingStart("cargo.recorder",base);recordErr:=record(base);finish(recordErr)
 timingStats(base,target,started)
 if recordErr!=nil{fmt.Fprintln(os.Stderr,recordErr)}
 if cargoErr!=nil{return commandStatus(cargoErr)}
 if recordErr!=nil||closeErr!=nil{return 96}
 return 0
}
`
