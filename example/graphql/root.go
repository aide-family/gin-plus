package main

import (
	"context"
	"embed"
	"errors"
)

//go:embed sdl
var content embed.FS

type Root struct{}

func (r *Root) Ping() string {
	return "pong"
}

type (
	AddArgs struct {
		A int32
		B int32
	}

	SumArgs struct {
		Nums []int32
	}

	SubArgs struct {
		A int32
		B int32
	}

	MultiArgs struct {
		A int32
		B int32
	}

	DivArgs struct {
		A int32
		B int32
	}

	Extend struct {
	}
)

func (r *Root) Add(ctx context.Context, args struct {
	In AddArgs
}) (int32, error) {
	return args.In.A + args.In.B, nil
}

func (r *Root) Sum(ctx context.Context, args struct {
	In SumArgs
}) (int32, error) {
	var sum int32
	for _, num := range args.In.Nums {
		sum += num
	}
	return sum, nil
}

func (r *Root) Sub(ctx context.Context, args struct {
	In SubArgs
}) (int32, error) {
	return args.In.A - args.In.B, nil
}

func (r *Root) Mul(ctx context.Context, args struct {
	In MultiArgs
}) (int32, error) {
	return args.In.A * args.In.B, nil
}

func (r *Root) Div(ctx context.Context, args struct {
	In DivArgs
}) (int32, error) {
	if args.In.B == 0 {
		return 0, errors.New("div by zero")
	}
	return args.In.A / args.In.B, nil
}

func (r *Root) CheckUUID(ctx context.Context, args struct {
	In UUID
}) (string, error) {
	return args.In.String(), nil
}

func (r *Root) Extend(ctx context.Context) *Extend {
	return &Extend{}
}

// Info ...
// request:
//
//	{
//	 extend {
//	   info
//	 }
//	}
//
// response:
//
//	{
//	 "data": {
//	   "extend": {
//	     "info": "extend info"
//	   }
//	 }
//	}
func (r *Extend) Info(ctx context.Context) string {
	return "extend info"
}
