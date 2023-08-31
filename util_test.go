package ginplus

import (
	"github.com/gin-gonic/gin"
	"testing"
)

type A struct {
}

type B struct {
}

func (a *A) Middlewares() []gin.HandlerFunc {
	return nil
}

func Test_isController(t *testing.T) {
	type args struct {
		c any
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test",
			args: args{
				c: &A{},
			},
			want: true,
		},
		{
			name: "test",
			args: args{
				c: &B{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := isMiddlewarer(tt.args.c)
			if got != tt.want {
				t.Errorf("isController() = %v, want %v", got, tt.want)
			}
		})
	}
}
