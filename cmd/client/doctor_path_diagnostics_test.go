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
