package shell2exec

import (
 "errors"
 "fmt"
 "os"
 "os/exec"
 "path/filepath"
 "runtime"
 "strings"
)

// Active reports an explicitly identified MSYS2 environment on Windows.
func Active() bool {
 return runtime.GOOS == "windows" && (os.Getenv("MSYSTEM") != "" || os.Getenv("MSYS2_ROOT") != "")
}

// scriptArgument translates a Windows path under MSYS2_ROOT into the
// POSIX namespace expected by the MSYS2 bash interpreter.
func scriptArgument(root, script string) string {
 rel, err := filepath.Rel(root, script)
 if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
  return "/" + filepath.ToSlash(rel)
 }
 return script
}

// Resolve returns the process command without running it. Native commands keep
// Go's usual lookup rules; MSYS2 extensionless scripts use their interpreter.
func Resolve(args []string) (*exec.Cmd, error) {
 if len(args) == 0 { return nil, errors.New("empty stdio command") }
 if termuxActive() { return resolveTermux(args) }
 if !Active() { return exec.Command(args[0], args[1:]...), nil }
 root := strings.TrimSpace(os.Getenv("MSYS2_ROOT"))
 if root == "" { return nil, errors.New("MSYS2 environment detected but MSYS2_ROOT is not set; fix: export MSYS2_ROOT=\"$(cygpath.exe -w /)\"") }
 if !filepath.IsAbs(root) { return nil, fmt.Errorf("MSYS2_ROOT must be an absolute Windows path: %q", root) }
 bash := filepath.Join(root, "usr", "bin", "bash.exe")
 if info, err := os.Stat(bash); err != nil || info.IsDir() {
  return nil, fmt.Errorf("invalid MSYS2_ROOT %q: usr/bin/bash.exe not found", root)
 }
 native := exec.Command(args[0], args[1:]...)
 if native.Err == nil { return native, nil }
 // Only attempt an extensionless script after native lookup has failed.
 if filepath.Ext(args[0]) != "" { return native, nil }
 names := []string{}
 if filepath.IsAbs(args[0]) || strings.ContainsAny(args[0], `/\`) { names = append(names, args[0]) } else {
  for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
   if dir != "" { names = append(names, filepath.Join(dir, args[0])) }
  }
 }
 for _, name := range names {
  info, err := os.Stat(name)
  if err != nil || info.IsDir() { continue }
  f, err := os.Open(name)
  if err != nil { return nil, err }
  b := make([]byte, 4096)
  n, _ := f.Read(b)
  _ = f.Close()
  if n < 2 || string(b[:2]) != "#!" { continue }
  line := strings.TrimSpace(strings.SplitN(string(b[2:n]), "\n", 2)[0])
  fields := strings.Fields(line)
  if len(fields) == 0 { continue }
  interpreter := strings.ReplaceAll(fields[0], "\\", "/")
  switch interpreter {
  case "/usr/bin/bash", "/bin/bash":
   return exec.Command(bash, append([]string{scriptArgument(root, name)}, args[1:]...)...), nil
  case "/usr/bin/env", "/bin/env":
   if len(fields) >= 2 && fields[1] == "bash" {
    return exec.Command(bash, append([]string{name}, args[1:]...)...), nil
   }
  }
 }
 return native, nil
}

 
// termuxActive requires Android and Termux's conventional environment.
func termuxActive() bool {
 return runtime.GOOS == "android" && os.Getenv("PREFIX") != "" && (os.Getenv("TERMUX_VERSION") != "" || strings.Contains(os.Getenv("PREFIX"), "/com.termux/"))
}

// resolveTermux bypasses incompatible /usr/bin/env and /bin/bash shebangs.
// Other scripts and native executables retain normal exec semantics.
func resolveTermux(args []string) (*exec.Cmd, error) {
 native := exec.Command(args[0], args[1:]...)
 prefix := strings.TrimSpace(os.Getenv("PREFIX"))
 if !filepath.IsAbs(prefix) { return nil, fmt.Errorf("Termux PREFIX must be absolute: %q", prefix) }
 names := []string{}
 if filepath.IsAbs(args[0]) || strings.ContainsRune(args[0], filepath.Separator) {
  names = append(names, args[0])
 } else {
  for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
   if dir != "" { names = append(names, filepath.Join(dir, args[0])) }
  }
 }
 for _, name := range names {
  info, err := os.Stat(name)
  if err != nil || info.IsDir() { continue }
  f, err := os.Open(name)
  if err != nil { return nil, err }
  buf := make([]byte, 4096)
  n, _ := f.Read(buf)
  _ = f.Close()
  if n < 2 || string(buf[:2]) != "#!" { return native, nil }
  fields := strings.Fields(strings.TrimSpace(strings.SplitN(string(buf[2:n]), "\n", 2)[0]))
  if len(fields) == 0 { return native, nil }
  interpreter := fields[0]
  isBash := interpreter == "/bin/bash" || interpreter == "/usr/bin/bash"
  isEnvBash := (interpreter == "/usr/bin/env" || interpreter == "/bin/env") && len(fields) >= 2 && (fields[1] == "bash" || (len(fields) >= 3 && fields[1] == "-S" && fields[2] == "bash"))
  if !isBash && !isEnvBash { return native, nil }
  bash := filepath.Join(prefix, "bin", "bash")
  if stat, err := os.Stat(bash); err != nil || stat.IsDir() { return nil, fmt.Errorf("Termux bash not found at %q", bash) }
  return exec.Command(bash, append([]string{name}, args[1:]...)...), nil
 }
 return native, nil
}
