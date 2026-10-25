package msys2exec

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

// Resolve returns the process command without running it. Native commands keep
// Go's usual lookup rules; MSYS2 extensionless scripts use their interpreter.
func Resolve(args []string) (*exec.Cmd, error) {
 if len(args) == 0 { return nil, errors.New("empty stdio command") }
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
   return exec.Command(bash, append([]string{name}, args[1:]...)...), nil
  case "/usr/bin/env", "/bin/env":
   if len(fields) >= 2 && fields[1] == "bash" {
    return exec.Command(bash, append([]string{name}, args[1:]...)...), nil
   }
  }
 }
 return native, nil
}
