package dataplane

import (
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/vishvananda/netlink"
)

func Test_getVtepByVNI(t *testing.T) {
	type args struct {
		vni uint32
	}
	tests := []struct {
		name    string
		args    args
		want    netlink.Link
		wantErr bool
	}{
		{
			name: "not exist",
			args: args{
				vni: 123,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "not match",
			args: args{
				vni: 123,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "normal", // HACK(shawn): Run go test with gcflags to disable inline 'go test -gcflags=all=-l'
			args: args{
				vni: 123,
			},
			want:    &netlink.Vxlan{VxlanId: 123},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "not exist" {
				linkByNamePatch := gomonkey.ApplyFunc(netlink.LinkByName,
					func(_ string) (netlink.Link, error) { return nil, errors.New("error") })
				defer linkByNamePatch.Reset()
			} else if tt.name == "not match" {
				linkByNamePatch := gomonkey.ApplyFunc(netlink.LinkByName,
					func(_ string) (netlink.Link, error) { return nil, errors.New("error") })
				defer linkByNamePatch.Reset()
			} else if tt.name == "normal" {
				linkByNamePatch := gomonkey.ApplyFunc(netlink.LinkByName,
					func(_ string) (netlink.Link, error) { return &netlink.Vxlan{VxlanId: 123}, nil })
				defer linkByNamePatch.Reset()
			}
			got, err := getVtepByVNI(tt.args.vni)
			if (err != nil) != tt.wantErr {
				t.Errorf("getVtepByVNI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getVtepByVNI() = %v, want %v", got, tt.want)
			}
		})
	}
}
