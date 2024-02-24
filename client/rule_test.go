package client

import "testing"

func TestRuleManager_Get(t *testing.T) {
	r, err := NewRuleManager([]string{
		"host-suffix: ad.com, reject",
		"host-prefix: ad.com, reject",
		"host-match: cdn, direct",
		"host-regexp: www\\.[a-zA-Z]+\\.com, direct",
		"ip-cidr: 127.0.0.1/8, direct",
		"geo: CN, direct",
		"match-all, forward: local",
	})
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		host string
	}
	tests := []struct {
		name       string
		args       args
		wantServer string
	}{
		{
			name: "1",
			args: args{
				host: "www.ad.com",
			},
			wantServer: "reject",
		},
		{
			name: "2",
			args: args{
				host: "www.baidu.com",
			},
			wantServer: "direct",
		},
		{
			name: "3",
			args: args{
				host: "ad.com",
			},
			wantServer: "reject",
		},
		{
			name: "4",
			args: args{
				host: "cdn.com",
			},
			wantServer: "direct",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotServer := r.Get(tt.args.host); gotServer != tt.wantServer {
				t.Errorf("Get() = %v, want %v", gotServer, tt.wantServer)
			}
		})
	}
}
