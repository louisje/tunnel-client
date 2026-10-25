package main

import (
 "os"
 "path/filepath"
 "runtime"
 "strings"
 "testing"
)

func TestDoctorStdioExecutableEvidenceMissing(t *testing.T) {
 evidence := doctorStdioExecutableEvidence("tunnel_client_nonexistent_executable_7852 --flag")
 output := strings.Join(evidence, "\n")
 if !strings.Contains(output, "host OS: "+runtime.GOOS) {
  t.Fatalf("missing OS diagnostic: %s", output)
 }
 if !strings.Contains(output, "process PATH entries:") {
  t.Fatalf("missing PATH summary: %s", output)
 }
 if !strings.Contains(output, "native executable lookup:") {
  t.Fatalf("missing native lookup: %s", output)
 }
}

func TestDoctorStdioExecutableEvidenceWindowsScript(t *testing.T) {
 if runtime.GOOS != "windows" {
  t.Skip("Windows-specific PATH lookup")
 }
 dir := t.TempDir()
 path := filepath.Join(dir, "bunpx")
 if err := os.WriteFile(path, []byte("#!/usr/bin/env bash\necho test\n"), 0600); err != nil {
  t.Fatal(err)
 }
 t.Setenv("PATH", dir)
 evidence := strings.Join(doctorStdioExecutableEvidence("bunpx --version"), "\n")
 if !strings.Contains(evidence, "file exists in process PATH: "+path) {
  t.Fatalf("missing candidate: %s", evidence)
 }
 if !strings.Contains(evidence, "script shebang: /usr/bin/env bash") {
  t.Fatalf("missing shebang: %s", evidence)
 }
}

func TestDoctorMSYS2InterpreterAbsoluteBash(t *testing.T) {
 if runtime.GOOS != "windows" { t.Skip("Windows-specific") }
 root := t.TempDir()
 t.Setenv("MSYSTEM", "MSYS")
 t.Setenv("MSYS2_ROOT", root)
 bin := filepath.Join(root, "usr", "bin")
 if err := os.MkdirAll(bin, 0755); err != nil { t.Fatal(err) }
 bash := filepath.Join(bin, "bash.exe")
 if err := os.WriteFile(bash, []byte("fake exe"), 0600); err != nil { t.Fatal(err) }
 script := filepath.Join(root, "home", "user", "outland", "bin", "bunpx")
 lines := strings.Join(doctorMSYS2Interpreter(script, "/usr/bin/bash", nil), "\n")
 if !strings.Contains(lines, "MSYS2 root: "+root) || !strings.Contains(lines, "resolved shebang interpreter: "+bash) {
  t.Fatalf("unexpected resolution: %s", lines)
 }
}

func TestDoctorMSYS2InterpreterEnvBashFromPath(t *testing.T) {
 if runtime.GOOS != "windows" { t.Skip("Windows-specific") }
 root := t.TempDir()
 t.Setenv("MSYSTEM", "MSYS")
 t.Setenv("MSYS2_ROOT", root)
 bin := filepath.Join(root, "usr", "bin")
 if err := os.MkdirAll(bin, 0755); err != nil { t.Fatal(err) }
 for _, file := range []string{"bash.exe", "env.exe"} {
  if err := os.WriteFile(filepath.Join(bin, file), []byte("fake exe"), 0600); err != nil { t.Fatal(err) }
 }
 lines := strings.Join(doctorMSYS2Interpreter(filepath.Join(t.TempDir(), "bunpx"), "/usr/bin/env bash", []string{bin}), "\n")
 if !strings.Contains(lines, "resolved shebang interpreter: "+filepath.Join(bin, "env.exe")) || !strings.Contains(lines, "resolved env bash: "+filepath.Join(bin, "bash.exe")) {
  t.Fatalf("unexpected env resolution: %s", lines)
 }
}
