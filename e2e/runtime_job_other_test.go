//go:build !windows

package e2e_test

func runtimeArtifactJob(int) (func(), error) { return func() {}, nil }
