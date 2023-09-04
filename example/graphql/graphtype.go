package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go/decode"
)

type UUID string

// 自定义类型实现graphql类型接口
var _ decode.Unmarshaler = (*UUID)(nil)
var _ fmt.Stringer = (*UUID)(nil)

func (*UUID) ImplementsGraphQLType(name string) bool {
	return name == "UUID"
}

func (U *UUID) UnmarshalGraphQL(input interface{}) error {
	val, ok := input.(string)
	if !ok {
		return fmt.Errorf("[UUID] wrong type, input: %T", input)
	}

	// 验证是否是uuid
	if _, err := uuid.Parse(val); err != nil {
		return fmt.Errorf("[UUID] wrong format, input: %s", val)
	}

	*U = UUID(val)
	return nil
}

func (U *UUID) String() string {
	return string(*U)
}
