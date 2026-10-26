package native

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

var controlledCargoOnce sync.Once
var controlledCargoDirectory, controlledCargoPath string
var controlledCargoError error

func TestMain(m *testing.M) {
	status := m.Run()
	if controlledCargoDirectory != "" {
		if err := os.RemoveAll(controlledCargoDirectory); err != nil {
			fmt.Fprintln(os.Stderr, "clean controlled Cargo fixture:", err)
			status = 1
		}
	}
	os.Exit(status)
}

// The immutable helper implements only the observer's test protocol. Its bytes
// are never accepted as real Cargo provenance by a frontend/runtime parity test.
// Independent environment inputs retain every invocation and failure scenario.
func controlledCargo(t testing.TB) string {
	t.Helper()
	controlledCargoOnce.Do(func() {
		controlledCargoDirectory, controlledCargoError = os.MkdirTemp("", "l-controlled-cargo-")
		if controlledCargoError != nil {
			return
		}
		source := filepath.Join(controlledCargoDirectory, "main.go")
		controlledCargoPath = filepath.Join(controlledCargoDirectory, "controlled-cargo")
		if controlledCargoError = os.WriteFile(source, []byte(controlledCargoSource), 0600); controlledCargoError != nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), BuildPrerequisiteTimeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, "go", "build", "-o", controlledCargoPath, source)
		cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off", "GOENV=off", "GOFLAGS=")
		if output, err := cmd.CombinedOutput(); err != nil {
			controlledCargoError = fmt.Errorf("compile controlled Cargo fixture: %w: %s", err, output)
		}
	})
	if controlledCargoError != nil {
		t.Fatal(controlledCargoError)
	}
	return controlledCargoPath
}

const controlledCargoSource = `package main
import("encoding/json";"fmt";"io";"os";"os/signal";"path/filepath";"strconv";"strings";"syscall";"time")
func write(path string,data []byte){if err:=os.WriteFile(path,data,0600);err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(99)}}
func terminate(){signal.Reset(syscall.SIGTERM);if err:=syscall.Kill(os.Getpid(),syscall.SIGTERM);err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(99)};time.Sleep(time.Second);os.Exit(99)}
func status(name string)int{value:=os.Getenv(name);if value==""{return 0};n,err:=strconv.Atoi(value);if err!=nil{fmt.Fprintln(os.Stderr,err);return 99};return n}
func main(){
 args:=os.Args[1:];metadata:=len(args)>0&&args[0]=="metadata"
 switch os.Getenv("L_CONTROLLED_CARGO_MODE") {
 case "helper":
  if len(args)>0&&args[0]=="which"{path,err:=os.Executable();if err!=nil{os.Exit(99)};fmt.Println(path);return}
  if len(args)>0&&args[0]=="-vV"{fmt.Print("release: 1.98.1\nhost: controlled-helper-host\n");return}
  emit:=func(name string,dest io.Writer){if path:=os.Getenv(name);path!=""{data,err:=os.ReadFile(path);if err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(99)};if _,err=dest.Write(data);err!=nil{os.Exit(99)}}}
  if len(args)>0&&args[0]=="locate-project"{
   write(filepath.Join(os.Getenv("L_HELPER_ROOT"),"locate-entered"),[]byte("entered"))
   if os.Getenv("L_HELPER_LOCATE_STDOUT")==""{fmt.Println(filepath.Join(os.Getenv("L_HELPER_WORKSPACE"),"Cargo.toml"))}else{emit("L_HELPER_LOCATE_STDOUT",os.Stdout)}
   emit("L_HELPER_LOCATE_STDERR",os.Stderr);os.Exit(status("L_HELPER_LOCATE_STATUS"))
  }
  manifest:="";for i,arg:=range args{if arg=="--manifest-path"&&i+1<len(args){manifest=args[i+1]}}
  if manifest==""{fmt.Fprintln(os.Stderr,"missing helper manifest");os.Exit(99)}
  root:=os.Getenv("L_HELPER_ROOT");write(filepath.Join(root,"build-entered"),[]byte(manifest))
  if release:=os.Getenv("L_HELPER_RELEASE");release!=""{until:=time.Now().Add(5*time.Second);for{if _,err:=os.Stat(release);err==nil{break};if time.Now().After(until){fmt.Fprintln(os.Stderr,"controlled Cargo release watchdog");os.Exit(99)};time.Sleep(time.Millisecond)}}
  data,err:=os.ReadFile(filepath.Join(filepath.Dir(manifest),"src","main.rs"));if err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(99)};write(filepath.Join(root,"observed-source"),data)
  emit("L_HELPER_BUILD_STDOUT",os.Stdout);emit("L_HELPER_BUILD_STDERR",os.Stderr);os.Exit(status("L_HELPER_BUILD_STATUS"))
 case "proxy":
  root:=os.Getenv("L_PROXY_ROOT");workspace:=os.Getenv("L_PROXY_WORKSPACE")
  if len(args)>0&&args[0]=="locate-project"{fmt.Println(filepath.Join(workspace,"Cargo.toml"));return}
  if metadata{
   write(filepath.Join(root,"metadata.args"),[]byte(strings.Join(args,"\x00")+"\x00"))
   if os.Getenv("L_PROXY_METADATA_ERROR")=="yes"{fmt.Fprint(os.Stderr,"metadata failure");os.Exit(8)}
   _=json.NewEncoder(os.Stdout).Encode(map[string]string{"workspace_root":workspace});if os.Getenv("L_PROXY_METADATA_ERROR")=="valid-json"{os.Exit(8)};return
  }
  write(filepath.Join(root,"native.args"),[]byte(strings.Join(args,"\x00")+"\x00"))
  cwd,err:=os.Getwd();if err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(99)}
  write(filepath.Join(root,"native.cwd"),[]byte(cwd+"\n"));write(filepath.Join(root,"native.env"),[]byte(os.Getenv("L_PROXY_VALUE")))
  if suffix:=os.Getenv("L_PROXY_METADATA_DIRECTORY");suffix!=""{paths,err:=filepath.Glob(filepath.Join(root,"cargo.*.cwd"));if err!=nil{os.Exit(99)};for _,path:=range paths{if err=os.Mkdir(strings.TrimSuffix(path,".cwd")+"."+suffix,0700);err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(99)}}}
  _=json.NewEncoder(os.Stdout).Encode(map[string]any{"reason":"compiler-artifact","executable":os.Getenv("L_PROXY_ARTIFACT"),"target":map[string]any{"kind":[]string{"bin"}},"fresh":false})
  if _,err=io.Copy(os.Stdout,os.Stdin);err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(99)}
  fmt.Fprint(os.Stderr,"native stderr\x00tail\n");if os.Getenv("L_PROXY_SIGNAL")=="yes"{terminate()};os.Exit(status("L_PROXY_STATUS"))
 case "timing":
  root:=os.Getenv("L_TIMING_TEST_ROOT")
  if metadata{_=json.NewEncoder(os.Stdout).Encode(map[string]string{"workspace_root":filepath.Join(root,"workspace")});return}
  write(filepath.Join(root,"args"),[]byte(strings.Join(args,"\n")+"\n"))
  dir:=filepath.Join(root,"target","cargo-timings");if err:=os.MkdirAll(dir,0700);err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(99)};write(filepath.Join(dir,"cargo-timing.html"),[]byte("independent timing html"))
  fmt.Print("{\"reason\":\"compiler-artifact\",\"fresh\":true}\n{\"reason\":\"compiler-artifact\",\"fresh\":false}\n");fmt.Fprint(os.Stderr,"native stderr\x00tail\n");os.Exit(status("L_TIMING_TEST_STATUS"))
 case "signal":
  if metadata{fmt.Println("{}");return};fmt.Fprint(os.Stderr,"native signal stderr\x00tail\n");terminate()
 case "stdin":
  if metadata{fmt.Println("{}");return};if _,err:=io.Copy(os.Stdout,os.Stdin);err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(99)}
 default:fmt.Fprintln(os.Stderr,"unknown controlled Cargo mode");os.Exit(99)
 }
}
`
