package cmdtrie

import (
	"fmt"
	"reflect"
	"strconv"
)

type command struct {
	fn reflect.Value
}

type trieNode struct {
	nexts map[string]*trieNode
	cmd   *command
}

type CommandTrie[TCtx any] struct {
	root trieNode
}

func NewCommandTrie[TCtx any]() CommandTrie[TCtx] {
	return CommandTrie[TCtx]{
		root: trieNode{
			nexts: make(map[string]*trieNode),
		},
	}
}

func (t *CommandTrie[TCtx]) Add(fn any, prefixWords ...string) error {
	typ := reflect.TypeOf(fn)
	if typ.Kind() != reflect.Func {
		return fmt.Errorf("fn must be a function, but get a %v", typ.Kind())
	}

	for i := 0; i < typ.NumIn(); i++ {
		argType := typ.In(i)
		if argType.Kind() == reflect.Array && i < typ.NumIn()-1 {
			return fmt.Errorf("array argument must at the last one")
		}
	}

	ptr := &t.root
	for _, word := range prefixWords {
		next, ok := ptr.nexts[word]
		if !ok {
			next = &trieNode{
				nexts: make(map[string]*trieNode),
			}
			ptr.nexts[word] = next
		}
		ptr = next
	}

	ptr.cmd = &command{
		fn: reflect.ValueOf(fn),
	}
	return nil
}

func (t *CommandTrie[TCtx]) Execute(ctx TCtx, cmdWords ...string) error {
	var cmd *command
	var argWords []string

	ptr := &t.root
	for i := 0; i < len(cmdWords); i++ {
		next, ok := ptr.nexts[cmdWords[i]]
		if !ok {
			break
		}
		if next != nil {
			cmd = next.cmd
			argWords = cmdWords[i+1:]
		}

		ptr = next
	}
	if cmd == nil {
		return fmt.Errorf("command not found")
	}

	fnType := cmd.fn.Type()

	// 最后一个参数如果是数组，那么可以少一个参数
	if len(argWords) < fnType.NumIn()-1 {
		return fmt.Errorf("no enough arguments for command")
	}

	var callArgs []reflect.Value

	// 数组参数只能是最后一个，所以先处理最后一个参数前的参数
	for i := 0; i < fnType.NumIn()-1; i++ {
		val, err := t.parseValue(argWords[i], fnType.In(i))
		if err != nil {
			return fmt.Errorf("cannot parse function argument at %d, err: %s", i, err.Error())
		}

		callArgs = append(callArgs, val)
	}

	if fnType.NumIn() > 0 {
		lastArgType := fnType.In(fnType.NumIn() - 1)
		lastArgWords := argWords[fnType.NumIn()-1:]
		lastArgTypeKind := lastArgType.Kind()

		var lastArg reflect.Value
		if lastArgTypeKind == reflect.Array || lastArgTypeKind == reflect.Slice {
			if lastArgType.Kind() == reflect.Array {
				lastArg = reflect.New(lastArgType)
			} else if lastArgType.Kind() == reflect.Slice {
				lastArg = reflect.MakeSlice(lastArgType, len(lastArgWords), len(lastArgWords))
			}

			for i := 0; i < len(lastArgWords); i++ {
				eleVal, err := t.parseValue(lastArgWords[i], lastArgType.Elem())
				if err != nil {
					return fmt.Errorf("cannot parse as array element, err: %s", err.Error())
				}
				lastArg.Index(i).Set(eleVal)
			}

		} else {
			if len(lastArgWords) == 0 {
				return fmt.Errorf("no enough arguments for command")
			}

			var err error
			lastArg, err = t.parseValue(lastArgWords[0], lastArgType)
			if err != nil {
				return fmt.Errorf("cannot parse function argument at %d, err: %s", fnType.NumIn()-1, err.Error())
			}
		}

		callArgs = append(callArgs, lastArg)
	}

	cmd.fn.Call(callArgs)
	return nil
}

func (t *CommandTrie[TCtx]) parseValue(word string, valueType reflect.Type) (reflect.Value, error) {
	valTypeKind := valueType.Kind()

	if valTypeKind == reflect.String {
		return reflect.ValueOf(word), nil
	}

	if reflect.Int <= valTypeKind && valTypeKind <= reflect.Int64 {
		i, err := strconv.ParseInt(word, 0, 64)
		if err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(i).Convert(valueType), nil
	}

	if reflect.Uint <= valTypeKind && valTypeKind <= reflect.Uint64 {
		i, err := strconv.ParseUint(word, 0, 64)
		if err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(i).Convert(valueType), nil
	}

	if reflect.Float32 <= valTypeKind && valTypeKind <= reflect.Float64 {
		i, err := strconv.ParseFloat(word, 64)
		if err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(i).Convert(valueType), nil
	}

	if valTypeKind == reflect.Bool {
		b, err := strconv.ParseBool(word)
		if err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(b), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot parse string as %s", valueType.Name())
}
