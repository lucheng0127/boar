package main

import (
	"os"
	"reflect"
	"syscall"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/lucheng0127/boar/api"
	"github.com/lucheng0127/boar/dataplane"
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

			patchDataplane := gomonkey.ApplyMethod(reflect.TypeOf(dataplane.NewDataplane()),
				"Serve", func(_ *dataplane.Dataplane) {})
			defer patchDataplane.Reset()
			patchApi := gomonkey.ApplyMethod(reflect.TypeOf(api.NewApiServer()),
				"Serve", func(_ *api.APIServer) {})
			defer patchApi.Reset()
			patchExist := gomonkey.ApplyFunc(os.Exit, func(_ int) {})
			defer patchExist.Reset()

			go func() {
				time.Sleep(time.Microsecond)
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			}()
			main()
		})
	}
}
