package native

const cargoProxyConfigEnvironment = "L_HARNESS_CARGO_CONFIG"

// The shared executable receives immutable, probe-local configuration through
// its environment because os.Executable resolves the cargo symlink. Real Cargo
// is an absolute pinned path, so invoking it cannot recurse through this proxy.
const artifactProxySource = `
var proxyEvidenceRunID,proxyEvidenceTest,proxyEvidenceHost string
func proxyCargo(args []string) int {
 configPath:=os.Getenv("L_HARNESS_CARGO_CONFIG")
 var config struct{Directory,Cargo,RunID,Test,Host string}
 data,err:=os.ReadFile(configPath)
 if err!=nil{fmt.Fprintln(os.Stderr,"read Cargo observer configuration:",err);return 96}
 if err=json.Unmarshal(data,&config);err!=nil||!filepath.IsAbs(config.Directory)||!filepath.IsAbs(config.Cargo){fmt.Fprintln(os.Stderr,"invalid Cargo observer configuration");return 96}
 f,err:=os.CreateTemp(config.Directory,"cargo.")
 if err!=nil{fmt.Fprintln(os.Stderr,err);return 96};base:=f.Name();if err=f.Close();err!=nil{fmt.Fprintln(os.Stderr,err);return 96}
 cwd,err:=os.Getwd();if err!=nil{fmt.Fprintln(os.Stderr,err);return 96}
 for suffix,value:=range map[string]string{".cwd":cwd+"\n",".args":strings.Join(args,"\n")+"\n"}{if err=os.WriteFile(base+suffix,[]byte(value),0600);err!=nil{fmt.Fprintln(os.Stderr,err);return 96}}
 native:=append([]string(nil),args...)
 if len(native)>0&&strings.HasPrefix(native[0],"+"){
  if native[0]!="+` + Version + `"{fmt.Fprintln(os.Stderr,"wrong Cargo toolchain selector");return 96}
  native=native[1:]
 }
 proxyEvidenceRunID=config.RunID;proxyEvidenceTest=config.Test;proxyEvidenceHost=config.Host
 return timedCargo(append([]string{base,config.Cargo},native...))
}
`
