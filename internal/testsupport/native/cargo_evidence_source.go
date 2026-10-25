package native

// This collector is compiled into the existing provenance recorder. Both the
// CLI proxy and the direct test helper call these exact functions. It never
// repairs source, resolves dependencies, executes artifacts, or owns Cargo's
// status. Failed/partial capture remains explicitly incomplete diagnostic data.
const artifactEvidenceSource = `
const evidenceMaxFiles=4096
const evidenceMaxBytes int64=64<<20
const evidenceMaxOutput int64=128<<20
type evidenceInput struct {
 Route,Directory,Workspace,Manifest,Target,RunID string
 Test,OriginalWorkspace,PreferredSlot,SelectedSlot,Toolchain,Host string
 Command,Environment []string
 PID,PPID int
}
type evidenceFile struct { Path,SHA256 string;Bytes int64;Mode uint32;ModifiedAt time.Time }
type evidenceState struct { Input evidenceInput;StartedAt time.Time;Errors []string;Files,OriginalFiles []evidenceFile;Digest,OriginalDigest string }
type evidenceCapture struct { Directory string;State evidenceState;stream *evidenceStream }
type evidenceStream struct { file *os.File;bytes int64;err error }
type evidenceBuffer struct { bytes.Buffer;limit int;err error }
func (b *evidenceBuffer) Write(p []byte)(int,error){n:=len(p);left:=b.limit-b.Len();if left<n{b.err=fmt.Errorf("bounded diagnostic output exceeded");if left>0{_,_=b.Buffer.Write(p[:left])};return n,nil};_,_=b.Buffer.Write(p);return n,nil}
func (w *evidenceStream) Write(p []byte)(int,error){n:=len(p);if w==nil{return n,nil};if w.err!=nil{return n,nil};if w.bytes+int64(n)>evidenceMaxOutput{w.err=fmt.Errorf("diagnostic stream exceeds %d bytes",evidenceMaxOutput);return n,nil};if w.file!=nil{_,w.err=w.file.Write(p)};w.bytes+=int64(n);return n,nil}
func evidenceOnce(path string,data []byte)error{f,err:=os.OpenFile(path,os.O_WRONLY|os.O_CREATE|os.O_EXCL,0600);if err!=nil{return err};_,err=f.Write(data);closeErr:=f.Close();if err!=nil{return err};return closeErr}
func evidenceJSON(path string,value any)error{b,err:=json.MarshalIndent(value,"","  ");if err!=nil{return err};return evidenceOnce(path,append(b,'\n'))}
func evidenceReadJSON(path string,value any)error{b,err:=os.ReadFile(path);if err!=nil{return err};return json.Unmarshal(b,value)}
func evidenceID(id string)(string,error){
 if id==""{id=os.Getenv("L_HARNESS_RUN_ID")};if id==""{var nonce [12]byte;if _,err:=rand.Read(nonce[:]);err!=nil{return "",err};id=fmt.Sprintf("standalone-%d-%x",os.Getppid(),nonce)}
 if id=="."||id==".."{return "",fmt.Errorf("invalid evidence run identity")};for _,c:=range id{if !(c>='a'&&c<='z'||c>='A'&&c<='Z'||c>='0'&&c<='9'||c=='_'||c=='-'||c=='.'){return "",fmt.Errorf("invalid evidence run identity")}};return id,nil
}
func evidenceEnvironment(env []string)[]map[string]string{
 // Only these compiler inputs are evidence. Do not enumerate arbitrary
 // environment names, credential-related Cargo settings, or their digests.
 allowed:=map[string]bool{"CARGO_BUILD_JOBS":true,"CARGO_HOME":true,"RUSTUP_HOME":true,"GOMAXPROCS":true,"L_HARNESS_TIMING_DIR":true,"L_HARNESS_CARGO_EVIDENCE_DIR":true,"PATH":true,"RUSTC":true,"RUSTC_WRAPPER":true,"RUSTC_WORKSPACE_WRAPPER":true,"RUSTFLAGS":true,"RUSTDOCFLAGS":true,"CARGO_ENCODED_RUSTFLAGS":true,"RUSTUP_TOOLCHAIN":true,"CARGO_BUILD_TARGET":true,"CARGO_TARGET_DIR":true,"CARGO_NET_OFFLINE":true,"CARGO_TERM_COLOR":true,"CARGO_PROFILE_DEV_DEBUG":true,"CARGO_PROFILE_TEST_DEBUG":true,"CC":true,"CXX":true,"AR":true,"CFLAGS":true,"CXXFLAGS":true,"LDFLAGS":true,"TMPDIR":true,"L_HARNESS_RUN_ID":true,"CARGO_LOG":true}
 result:=[]map[string]string{};for _,entry:=range env{parts:=strings.SplitN(entry,"=",2);if len(parts)!=2||!allowed[parts[0]]{continue};result=append(result,map[string]string{"Name":parts[0],"Value":parts[1],"SHA256":fmt.Sprintf("%x",sha256.Sum256([]byte(parts[1])))})};return result
}
func evidenceTree(root,target,dest string)([]evidenceFile,string,error){
 var entries []evidenceFile;var total int64;digest:=sha256.New()
 err:=filepath.WalkDir(root,func(path string,entry os.DirEntry,err error)error{
  if err!=nil{return err};rel,err:=filepath.Rel(root,path);if err!=nil{return err};if entry.IsDir(){if rel!="."&&(entry.Name()==".git"||path==target){return filepath.SkipDir};return nil}
  info,err:=entry.Info();if err!=nil{return err};if !info.Mode().IsRegular(){return fmt.Errorf("nonregular input %s",path)};if len(entries)>=evidenceMaxFiles||info.Size()>evidenceMaxBytes-total{return fmt.Errorf("input evidence exceeds bounded inventory")}
  data,err:=os.ReadFile(path);if err!=nil{return err};after,err:=os.Stat(path);if err!=nil||info.Size()!=int64(len(data))||after.Size()!=info.Size()||!after.ModTime().Equal(info.ModTime()){return fmt.Errorf("input changed during snapshot: %s",path)}
  rel=filepath.ToSlash(rel);sum:=fmt.Sprintf("%x",sha256.Sum256(data));entryData:=evidenceFile{Path:rel,SHA256:sum,Bytes:int64(len(data)),Mode:uint32(info.Mode().Perm()),ModifiedAt:info.ModTime().UTC()};entries=append(entries,entryData);total+=int64(len(data));fmt.Fprintf(digest,"%d:%s%d:%s",len(rel),rel,len(data),sum)
  if dest!=""{out:=filepath.Join(dest,filepath.FromSlash(rel));if err=os.MkdirAll(filepath.Dir(out),0700);err!=nil{return err};if err=evidenceOnce(out,data);err!=nil{return err}};return nil
 });return entries,fmt.Sprintf("%x",digest.Sum(nil)),err
}
func evidenceCacheSlot(input *evidenceInput,started time.Time)map[string]any{
 if input.Route!="cli-proxy"{return map[string]any{"Available":input.PreferredSlot!="","Source":"caller-owned diagnostic slot","PreferredSlot":input.PreferredSlot,"SelectedSlot":input.SelectedSlot}}
 input.PreferredSlot="unavailable";input.SelectedSlot="unavailable";unavailable:=func(reason string)map[string]any{return map[string]any{"Available":false,"Reason":reason,"Requires":"L_HARNESS_TIMING_DIR cache.slot event"}}
 root:=os.Getenv("L_HARNESS_TIMING_DIR");if root==""{return unavailable("CLI timing is disabled")}
 f,err:=os.Open(filepath.Join(root,fmt.Sprintf("%d.jsonl",input.PPID)));if err!=nil{return unavailable("parent timing stream unavailable")};defer f.Close();if info,err:=f.Stat();err!=nil||info.Size()>16<<20{return unavailable("parent timing stream exceeds bounded read")}
 scan:=bufio.NewScanner(io.LimitReader(f,16<<20));scan.Buffer(make([]byte,4096),1<<20);var matched map[string]any
 physical:=func(path string)string{if real,err:=filepath.EvalSymlinks(path);err==nil{return real};return filepath.Clean(path)}
 for scan.Scan(){var event map[string]any;if json.Unmarshal(scan.Bytes(),&event)!=nil||event["phase"]!="cache.slot"{continue};subject,ok:=event["subject"].(string);if !ok||physical(subject)!=physical(input.Directory){continue};fields,ok:=event["fields"].(map[string]any);if !ok{continue};target,ok:=fields["target"].(string);if !ok||physical(target)!=physical(input.Target){continue};stamp,ok:=event["started_at"].(string);if !ok{continue};at,err:=time.Parse(time.RFC3339Nano,stamp);if err!=nil||at.After(started){continue};if _,ok=fields["preferred_slot"];!ok{continue};if _,ok=fields["selected_slot"];!ok{continue};matched=event}
 if scan.Err()!=nil{return unavailable("parent timing stream incomplete")};if matched==nil{return unavailable("no matching parent cache.slot event")}
 fields:=matched["fields"].(map[string]any);input.PreferredSlot=fmt.Sprint(fields["preferred_slot"]);input.SelectedSlot=fmt.Sprint(fields["selected_slot"]);return map[string]any{"Available":true,"Source":"actual parent cache.slot event","Event":matched}
}
func evidenceBegin(input evidenceInput)*evidenceCapture{
 root:=os.Getenv("L_HARNESS_CARGO_EVIDENCE_DIR");if root==""{return nil};capture:=&evidenceCapture{State:evidenceState{Input:input,StartedAt:time.Now().UTC()}}
 id,err:=evidenceID(input.RunID);if err!=nil{timingWrite(map[string]any{"event":"record","phase":"cargo.evidence-error","error":err.Error()});return capture};capture.State.Input.RunID=id
 root,err=filepath.Abs(root);if err!=nil{return capture};run:=filepath.Join(root,id);if err=os.MkdirAll(filepath.Join(run,"invocations"),0700);err!=nil{return capture}
 runRecord:=map[string]any{"RunID":id,"Schema":1,"Purpose":"diagnostic-only unchanged Cargo inputs and results"};if err=evidenceJSON(filepath.Join(run,"run.json"),runRecord);err!=nil&&!os.IsExist(err){return capture}
 capture.Directory,err=os.MkdirTemp(filepath.Join(run,"invocations"),fmt.Sprintf("%d-",os.Getpid()));if err!=nil{return capture}
 environment:=input.Environment;if environment==nil{environment=os.Environ()};input.RunID=id;input.Environment=nil;cacheSlot:=evidenceCacheSlot(&input,capture.State.StartedAt);capture.State.Input=input
 start:=map[string]any{"Input":input,"Environment":evidenceEnvironment(environment),"StartedAt":capture.State.StartedAt,"RecorderPID":os.Getpid(),"RecorderPPID":os.Getppid(),"CacheSlotAttribution":cacheSlot}
 // Only the finite compiler-affecting environment allowlist is persisted.
 if err=evidenceJSON(filepath.Join(capture.Directory,"start.json"),start);err!=nil{capture.State.Errors=append(capture.State.Errors,err.Error())}
 if err=evidenceJSON(filepath.Join(capture.Directory,"cache-slot.json"),cacheSlot);err!=nil{capture.State.Errors=append(capture.State.Errors,err.Error())}
 if !filepath.IsAbs(input.Workspace)||!filepath.IsAbs(input.Manifest){capture.State.Errors=append(capture.State.Errors,"workspace and selected manifest must be absolute")}else if rel,err:=filepath.Rel(input.Workspace,input.Manifest);err!=nil||!filepath.IsLocal(rel){capture.State.Errors=append(capture.State.Errors,"selected manifest is outside captured workspace")}else if rel,err:=filepath.Rel(input.Workspace,root);err==nil&&filepath.IsLocal(rel){capture.State.Errors=append(capture.State.Errors,"evidence root must be outside Cargo input workspace")}else{
  capture.State.Files,capture.State.Digest,err=evidenceTree(input.Workspace,input.Target,filepath.Join(capture.Directory,"inputs"));if err!=nil{capture.State.Errors=append(capture.State.Errors,err.Error())}
 }
 if err=evidenceJSON(filepath.Join(capture.Directory,"inputs.json"),map[string]any{"Files":capture.State.Files,"Digest":capture.State.Digest});err!=nil{capture.State.Errors=append(capture.State.Errors,err.Error())}
 if filepath.IsAbs(input.OriginalWorkspace)&&input.OriginalWorkspace!=input.Workspace{capture.State.OriginalFiles,capture.State.OriginalDigest,err=evidenceTree(input.OriginalWorkspace,input.Target,"");if err!=nil{capture.State.Errors=append(capture.State.Errors,err.Error())};if err=evidenceJSON(filepath.Join(capture.Directory,"original-inputs.json"),map[string]any{"Root":input.OriginalWorkspace,"Files":capture.State.OriginalFiles,"Digest":capture.State.OriginalDigest});err!=nil{capture.State.Errors=append(capture.State.Errors,err.Error())}}
 if err=evidenceJSON(filepath.Join(capture.Directory,"state.json"),capture.State);err!=nil{capture.State.Errors=append(capture.State.Errors,err.Error())};return capture
}
func (c *evidenceCapture) fileStream(name string)*evidenceStream{file,err:=os.OpenFile(filepath.Join(c.Directory,name),os.O_WRONLY|os.O_CREATE|os.O_EXCL,0600);return &evidenceStream{file:file,err:err}}
func (c *evidenceCapture) stderr()io.Writer{if c==nil||c.Directory==""{return nil};c.stream=c.fileStream("cargo.stderr");return c.stream}
func evidenceLocate(cargo,manifest,cwd string)(string,map[string]any,error){
 ctx,cancel:=context.WithTimeout(context.Background(),25*time.Second);defer cancel()
 cmd:=exec.CommandContext(ctx,cargo,"locate-project","--workspace","--frozen","--message-format","plain","--manifest-path",manifest);cmd.Dir=cwd
 out,diag:=evidenceBuffer{limit:64<<10},evidenceBuffer{limit:1<<20};cmd.Stdout=&out;cmd.Stderr=&diag;cmd.WaitDelay=2*time.Second;finish:=timingStart("evidence.locate-project",manifest);err:=cmd.Run();finish(err);if err==nil{if out.err!=nil{err=out.err}else if diag.err!=nil{err=diag.err}}
 record:=map[string]any{"Command":cmd.Args,"Directory":cwd,"Stdout":out.String(),"Stderr":diag.String(),"Status":commandStatus(err)}
 if err!=nil{return "",record,err};workspace:=filepath.Dir(strings.TrimSpace(out.String()));if !filepath.IsAbs(workspace){return "",record,fmt.Errorf("Cargo locate-project returned no absolute workspace")};return workspace,record,nil
}
func evidenceCopy(source,dest string,limit int64)error{f,err:=os.Open(source);if err!=nil{return err};defer f.Close();info,err:=f.Stat();if err!=nil{return err};if !info.Mode().IsRegular()||info.Size()>limit{return fmt.Errorf("nonregular or oversized evidence %s",source)};data,err:=io.ReadAll(io.LimitReader(f,limit+1));if err!=nil{return err};if int64(len(data))>limit{return fmt.Errorf("oversized evidence %s",source)};return evidenceOnce(dest,data)}
func (c *evidenceCapture) finish(stdout,stderr []byte,status int,nativeError,base string){
 if c==nil||c.Directory==""{return};errors:=append([]string(nil),c.State.Errors...);add:=func(err error){if err!=nil{errors=append(errors,err.Error())}}
 if c.stream!=nil{if c.stream.file!=nil{add(c.stream.file.Close())};add(c.stream.err)}else if int64(len(stderr))>evidenceMaxOutput{add(fmt.Errorf("diagnostic stderr exceeds limit"))}else{add(evidenceOnce(filepath.Join(c.Directory,"cargo.stderr"),stderr))}
 if base!=""{add(evidenceCopy(base+".jsonl",filepath.Join(c.Directory,"cargo.stdout"),evidenceMaxOutput))}else if int64(len(stdout))>evidenceMaxOutput{add(fmt.Errorf("diagnostic stdout exceeds limit"))}else{add(evidenceOnce(filepath.Join(c.Directory,"cargo.stdout"),stdout))}
 var metadata []byte
 if base!=""{add(evidenceCopy(base+".metadata.json",filepath.Join(c.Directory,"metadata.json"),evidenceMaxOutput));add(evidenceCopy(base+".metadata.stderr",filepath.Join(c.Directory,"metadata.stderr"),evidenceMaxOutput));metadata,_=os.ReadFile(filepath.Join(c.Directory,"metadata.json"))}else if len(c.State.Input.Command)>0{
  cmd:=exec.Command(c.State.Input.Command[0],"metadata","--manifest-path",c.State.Input.Manifest,"--frozen","--offline","--no-deps","--format-version","1");cmd.Dir=c.State.Input.Directory
  out,diag:=c.fileStream("metadata.json"),c.fileStream("metadata.stderr");cmd.Stdout,cmd.Stderr=out,diag;finish:=timingStart("evidence.metadata",c.Directory);err:=cmd.Run();finish(err);add(err);for _,writer:=range []*evidenceStream{out,diag}{if writer.file!=nil{add(writer.file.Close())};add(writer.err)};metadata,_=os.ReadFile(filepath.Join(c.Directory,"metadata.json"))
 }
 var meta struct{WorkspaceRoot string ` + "`json:\"workspace_root\"`" + `};if err:=json.Unmarshal(metadata,&meta);err!=nil{add(err)}else{left,leftErr:=filepath.EvalSymlinks(meta.WorkspaceRoot);right,rightErr:=filepath.EvalSymlinks(c.State.Input.Workspace);if leftErr!=nil||rightErr!=nil||left!=right{add(fmt.Errorf("Cargo metadata workspace does not match captured root: %q versus %q",meta.WorkspaceRoot,c.State.Input.Workspace))}}
 var err error;verify:=func(label,root string,before []evidenceFile,digest string){if !filepath.IsAbs(root){return};after,afterDigest,treeErr:=evidenceTree(root,c.State.Input.Target,"");add(treeErr);if afterDigest!=digest{add(fmt.Errorf("%s input bytes changed during Cargo invocation",label));return};if len(before)!=len(after){add(fmt.Errorf("%s input inventory changed during Cargo invocation",label));return};for i,prior:=range before{current:=after[i];if prior.Mode!=current.Mode||!prior.ModifiedAt.Equal(current.ModifiedAt){add(fmt.Errorf("%s input mode/timestamp changed during Cargo invocation: %s",label,prior.Path))}}};verify("staged",c.State.Input.Workspace,c.State.Files,c.State.Digest);if c.State.OriginalDigest!=""{verify("original",c.State.Input.OriginalWorkspace,c.State.OriginalFiles,c.State.OriginalDigest)}
 fresh,rebuilt:=0,0;var artifactBytes int64;var artifacts []map[string]any;output,err:=os.Open(filepath.Join(c.Directory,"cargo.stdout"));if err==nil{scan:=bufio.NewScanner(output);scan.Buffer(make([]byte,4096),8*1024*1024);for scan.Scan(){var message struct{Reason,Executable string;Fresh bool;Target,Profile,Features json.RawMessage};if json.Unmarshal(scan.Bytes(),&message)!=nil||message.Reason!="compiler-artifact"{continue};if message.Fresh{fresh++}else{rebuilt++};if message.Executable!=""{info,statErr:=os.Stat(message.Executable);if statErr!=nil{add(statErr);continue};if !info.Mode().IsRegular()||info.Size()>evidenceMaxOutput-artifactBytes{add(fmt.Errorf("nonregular or oversized artifact %s",message.Executable));continue};data,readErr:=os.ReadFile(message.Executable);if readErr!=nil{add(readErr);continue};artifactBytes+=int64(len(data));sum:=fmt.Sprintf("%x",sha256.Sum256(data));dir:=filepath.Join(c.Directory,"artifacts");add(os.MkdirAll(dir,0700));writeErr:=evidenceOnce(filepath.Join(dir,sum),data);if writeErr!=nil&&!os.IsExist(writeErr){add(writeErr)};artifacts=append(artifacts,map[string]any{"Source":message.Executable,"SHA256":sum,"Bytes":len(data),"Target":message.Target,"Profile":message.Profile,"Features":message.Features})}};add(scan.Err());add(output.Close())}else{add(err)}
 report:=filepath.Join(c.State.Input.Target,"cargo-timings","cargo-timing.html");if info,err:=os.Stat(report);err==nil&&!info.ModTime().Before(c.State.StartedAt){add(evidenceCopy(report,filepath.Join(c.Directory,"cargo-timing.html"),evidenceMaxOutput))}
 complete:=map[string]any{"Complete":len(errors)==0,"Status":status,"NativeError":nativeError,"Errors":errors,"FinishedAt":time.Now().UTC(),"DurationNS":time.Since(c.State.StartedAt).Nanoseconds(),"Fresh":fresh,"Rebuilt":rebuilt,"Artifacts":artifacts,"InputDigest":c.State.Digest,"RunID":c.State.Input.RunID}
 if err=evidenceJSON(filepath.Join(c.Directory,"complete.json"),complete);err!=nil{complete["Complete"]=false;complete["Errors"]=append(errors,err.Error());timingWrite(map[string]any{"event":"record","phase":"cargo.evidence-error","subject":c.Directory,"error":err.Error()})};timingWrite(map[string]any{"event":"record","phase":"cargo.evidence","subject":c.Directory,"fields":complete})
}
func evidenceCommand(args []string)int{
 if len(args)==0{return 96};switch args[0]{case "__evidence_begin":var input evidenceInput;if err:=json.NewDecoder(os.Stdin).Decode(&input);err!=nil{fmt.Fprintln(os.Stderr,err);return 96};capture:=evidenceBegin(input);if capture==nil||capture.Directory==""{fmt.Fprintln(os.Stderr,"Cargo evidence initialization failed");return 96};fmt.Println(capture.Directory);return 0
 case "__evidence_finish":var request struct{Directory string;Stdout,Stderr []byte;Status int;Error string};if err:=json.NewDecoder(os.Stdin).Decode(&request);err!=nil{fmt.Fprintln(os.Stderr,err);return 96};capture:=&evidenceCapture{Directory:request.Directory};if err:=evidenceReadJSON(filepath.Join(request.Directory,"state.json"),&capture.State);err!=nil{fmt.Fprintln(os.Stderr,err);return 96};capture.finish(request.Stdout,request.Stderr,request.Status,request.Error,"");return 0};return 96
}
`
