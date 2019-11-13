// +build !skip_mysql

package mysql

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func Test_buildConnectionString(t *testing.T) {
	type args struct {
		databaseName string
	}

	tests := []struct {
		name    string
		args    args
		opts    mysqlOpts
		setOpts bool
		want    string
	}{
		{args: args{databaseName: "ssetest"}, setOpts: true, want: "tcp(192.0.2.0)/ssetest",
			opts: mysqlOpts{Host: "192.0.2.0", Database: "ssetest"},
		},
		{args: args{databaseName: "ssetest"}, setOpts: true, want: "/ssetest",
			opts: mysqlOpts{Port: "3307", Database: "ssetest"},
		},
		{args: args{databaseName: "ssetest"}, setOpts: true, want: "sseuser:passwd@tcp(192.0.2.0)/ssetest",
			opts: mysqlOpts{Host: "192.0.2.0", Database: "ssetest", User: "sseuser", Password: "passwd"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setOpts {
				opts = &tt.opts
			}
			if got := buildConnectionString(tt.args.databaseName); got != tt.want {
				t.Errorf("buildConnectionString() = %v, want %v", got, tt.want)
			}
		})
	}
}
