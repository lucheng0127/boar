package main

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "normal",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = append(os.Args, "--cfg=./conf/boar.yaml")
			os.Args = append(os.Args, "--debug")

			go func() {
				time.Sleep(5 * time.Second)
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			}()
			main()
		})
	}
}
