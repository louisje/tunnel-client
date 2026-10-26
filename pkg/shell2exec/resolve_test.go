package shell2exec

import (
 "os"
 "path/filepath"
 "runtime"
 "strings"
 "testing"
)

func TestNativeCommandUnchanged(t *testing.T) {
 cmd, err := Resolve([]string{"unavailable_mcp_command_test", "--test"})
 if err != nil { t.Fatal(err) }
 if cmd == nil { t.Fatal("nil command") }
}

func TestMSYS2RootRequired(t *testing.T) {
 if runtime.GOOS != "windows" { t.Skip("Windows-only") }
 t.Setenv("MSYSTEM", "MSYS")
 t.Setenv("MSYS2_ROOT", "")
 _, err := Resolve([]string{"bunpx"})
 if err == nil || !strings.Contains(err.Error(), "MSYS2_ROOT is not set") { t.Fatalf("expected missing root error, got %v", err) }
}

func TestMSYS2ScriptShebang(t *testing.T) {
 if runtime.GOOS != "windows" { t.Skip("Windows-only") }
 root := t.TempDir()
 bin := filepath.Join(root, "usr", "bin")
 if err := os.MkdirAll(bin, 0755); err != nil { t.Fatal(err) }
 if err := os.WriteFile(filepath.Join(bin, "bash.exe"), []byte("dummy"), 0600); err != nil { t.Fatal(err) }
 scripts := filepath.Join(root, "scripts")
 if err := os.MkdirAll(scripts, 0755); err != nil { t.Fatal(err) }
 script := filepath.Join(scripts, "bunpx")
 if err := os.WriteFile(script, []byte("#!/usr/bin/env bash\necho test\n"), 0600); err != nil { t.Fatal(err) }
 t.Setenv("MSYSTEM", "MSYS")
 t.Setenv("MSYS2_ROOT", root)
 t.Setenv("PATH", scripts)
 cmd, err := Resolve([]string{"bunpx", "hello"})
 if err != nil { t.Fatal(err) }
 if cmd.Path != filepath.Join(bin, "bash.exe") { t.Fatalf("unexpected executable: %s", cmd.Path) }
 if len(cmd.Args) != 3 || cmd.Args[1] != "/scripts/bunpx" || cmd.Args[2] != "hello" { t.Fatalf("unexpected argv: %#v", cmd.Args) }
}

func TestTermuxScriptRecognition(t *testing.T) {
 if runtime.GOOS != "android" { t.Skip("Android-only") }
 prefix := t.TempDir()
 if err := os.MkdirAll(filepath.Join(prefix,"bin"),0755); err != nil { t.Fatal(err) }
 if err := os.WriteFile(filepath.Join(prefix,"bin","bash"),[]byte("dummy"),0755); err != nil { t.Fatal(err) }
 script := filepath.Join(t.TempDir(),"bunpx")
 if err := os.WriteFile(script,[]byte("#!/usr/bin/env bash\necho test\n"),0755); err != nil { t.Fatal(err) }
 t.Setenv("PREFIX",prefix)
 cmd,err := resolveTermux([]string{script,"hello"})
 if err != nil { t.Fatal(err) }
 if cmd.Path != filepath.Join(prefix,"bin","bash") || cmd.Args[1] != script || cmd.Args[2] != "hello" { t.Fatalf("unexpected command: %#v",cmd.Args) }
}
