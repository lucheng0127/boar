package main

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Normal",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = append(os.Args, "-f=conf/boar.yaml")
			os.Args = append(os.Args, "-t=yaml")

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			go func() {
				time.Sleep(time.Microsecond)
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			}()
			main()
		})
	}
}
