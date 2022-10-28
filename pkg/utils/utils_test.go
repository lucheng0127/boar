package utils

import (
	"net"
	"reflect"
	"testing"
)

func TestParseNetworkInfo(t *testing.T) {
	type args struct {
		ip        net.IP
		subnetLen int
		netLen    int
	}
	tests := []struct {
		name    string
		args    args
		want    *net.IPNet
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				ip:        net.ParseIP("172.17.255.254"),
				subnetLen: 24,
				netLen:    16,
			},
			want: &net.IPNet{
				IP:   net.ParseIP("0.0.0.0"),
				Mask: net.CIDRMask(0, 32),
			},
		},
		{
			name: "18",
			args: args{
				ip:        net.ParseIP("172.17.254.65"),
				subnetLen: 18,
				netLen:    16,
			},
			want: &net.IPNet{
				IP:   net.ParseIP("172.17.64.0").To4(),
				Mask: net.CIDRMask(18, 32),
			},
		},
		{
			name: "22",
			args: args{
				ip:        net.ParseIP("192.168.254.5"),
				subnetLen: 22,
				netLen:    16,
			},
			want: &net.IPNet{
				IP:   net.ParseIP("192.168.4.0").To4(),
				Mask: net.CIDRMask(22, 32),
			},
		},
		{
			name: "24",
			args: args{
				ip:        net.ParseIP("172.17.254.124"),
				subnetLen: 24,
				netLen:    16,
			},
			want: &net.IPNet{
				IP:   net.ParseIP("172.17.123.0").To4(),
				Mask: net.CIDRMask(24, 32),
			},
		},
		{
			name: "20",
			args: args{
				ip:        net.ParseIP("192.168.142.135"),
				subnetLen: 24,
				netLen:    20,
			},
			want: &net.IPNet{
				IP:   net.ParseIP("192.168.134.0").To4(),
				Mask: net.CIDRMask(24, 32),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseNetworkInfo(tt.args.ip, tt.args.subnetLen, tt.args.subnetLen)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNetworkInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseNetworkInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
