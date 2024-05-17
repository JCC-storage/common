package cmdtrie

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/samber/lo"
	"gitlink.org.cn/cloudream/common/pkgs/trie"
	"gitlink.org.cn/cloudream/common/utils/reflect2"
)

var ErrCommandNotFound = fmt.Errorf("command not found")

type ExecuteOption struct {
	ReplaceEmptyArrayWithNil bool // 如果最后一个参数是空数组，则调用命令的时候传递nil参数
}

type command struct {
	fn             reflect.Value
	fnType         reflect.Type
	staticArgTypes []reflect.Type
	lastIsArray    bool
}

type anyCommandTrie struct {
	trie    trie.Trie[*command]
	ctxType reflect.Type
	retType reflect.Type
}

func newAnyCommandTrie(ctxType reflect.Type, retType reflect.Type) anyCommandTrie {
	return anyCommandTrie{
		ctxType: ctxType,
		retType: retType,
	}
}

func (t *anyCommandTrie) Add(fn any, prefixWords ...string) error {
	typ := reflect.TypeOf(fn)
	if typ.Kind() != reflect.Func {
		return fmt.Errorf("fn must be a function, but get a %v", typ.Kind())
	}

	err := t.checkFnReturn(typ)
	if err != nil {
		return err
	}

	err = t.checkFnArgs(typ)
	if err != nil {
		return err
	}

	node := t.trie.Create(lo.Map(prefixWords, func(val string, index int) any { return val }))

	fnType := reflect.TypeOf(fn)
	var staticArgTypes []reflect.Type
	if t.ctxType != nil {
		for i := 1; i < fnType.NumIn(); i++ {
			staticArgTypes = append(staticArgTypes, fnType.In(i))
		}
	} else {
		for i := 0; i < fnType.NumIn(); i++ {
			staticArgTypes = append(staticArgTypes, fnType.In(i))
		}
	}

	var lastIsArray = false
	if len(staticArgTypes) > 0 {
		kind := staticArgTypes[len(staticArgTypes)-1].Kind()
		lastIsArray = kind == reflect.Array || kind == reflect.Slice
	}

	node.Value = &command{
		fn:             reflect.ValueOf(fn),
		fnType:         reflect.TypeOf(fn),
		staticArgTypes: staticArgTypes,
		lastIsArray:    lastIsArray,
	}
	return nil
}

func (t *anyCommandTrie) checkFnReturn(typ reflect.Type) error {
	if t.retType != nil {
		if typ.NumOut() != 1 {
			return fmt.Errorf("fn must have one return value with type %s", t.retType.Name())
		}

		fnRetType := typ.Out(0)
		if t.retType.Kind() == reflect.Interface {

			// 如果TRet是接口类型，那么fn的返回值只要实现了此接口，就也可以接受
			if !fnRetType.Implements(t.retType) {
				return fmt.Errorf("fn must have one return value with type %s", t.retType.Name())
			}

		} else if fnRetType != t.retType {
			return fmt.Errorf("fn must have one return value with type %s", t.retType.Name())
		}
	}
	return nil
}

func (t *anyCommandTrie) checkFnArgs(typ reflect.Type) error {
	if t.ctxType != nil {
		if typ.NumIn() < 1 {
			return fmt.Errorf("fn must have a ctx argument")
		}

		for i := 0; i < typ.NumIn(); i++ {
			argType := typ.In(i)
			if i == 0 && argType != t.ctxType {
				return fmt.Errorf("first argument of fn must be %s", t.ctxType.Name())
			}

			if argType.Kind() == reflect.Array && i < typ.NumIn()-1 {
				return fmt.Errorf("array argument must at the last one")
			}
		}
	} else {
		for i := 0; i < typ.NumIn(); i++ {
			argType := typ.In(i)
			if argType.Kind() == reflect.Array && i < typ.NumIn()-1 {
				return fmt.Errorf("array argument must at the last one")
			}
		}
	}
	return nil
}

func (t *anyCommandTrie) Execute(ctx any, cmdWords []string, opt ExecuteOption) ([]reflect.Value, error) {
	var cmd *command
	var argWords []string

	cmd, argWords, err := t.findCommand(cmdWords, argWords)
	if err != nil {
		return nil, err
	}

	if cmd.lastIsArray {
		// 最后一个参数如果是数组，那么可以少一个参数
		if len(argWords) < len(cmd.staticArgTypes)-1 {
			return nil, fmt.Errorf("no enough arguments for command")
		}
	} else if len(argWords) < len(cmd.staticArgTypes) {
		return nil, fmt.Errorf("no enough arguments for command")
	}

	var callArgs []reflect.Value

	// 如果有Ctx参数，则加上Ctx参数
	if t.ctxType != nil {
		callArgs = append(callArgs, reflect.ValueOf(ctx))
	}

	// 数组参数只能是最后一个，所以先处理最后一个参数前的参数
	callArgs, err = t.parseFrontArgs(cmd, argWords, callArgs)
	if err != nil {
		return nil, err
	}

	// 解析最后一个参数
	callArgs, err = t.parseLastArg(cmd, argWords, opt, callArgs)
	if err != nil {
		return nil, err
	}

	return cmd.fn.Call(callArgs), nil
}

func (t *anyCommandTrie) findCommand(cmdWords []string, argWords []string) (*command, []string, error) {
	var cmd *command

	t.trie.Walk(cmdWords, func(word string, index int, node *trie.Node[*command], isWordNode bool) {
		if node.Value != nil {
			cmd = node.Value
			argWords = cmdWords[index+1:]
		}
	})

	if cmd == nil {
		return nil, nil, ErrCommandNotFound
	}
	return cmd, argWords, nil
}

func (t *anyCommandTrie) parseFrontArgs(cmd *command, argWords []string, callArgs []reflect.Value) ([]reflect.Value, error) {
	for i := 0; i < len(cmd.staticArgTypes)-1; i++ {
		val, err := t.parseValue(argWords[i], cmd.staticArgTypes[i])
		if err != nil {
			// 如果有Ctx参数，则参数的位置要往后一个
			argIndex := i
			if t.ctxType != nil {
				argIndex++
			}

			return nil, fmt.Errorf("cannot parse function argument at %d, err: %s", argIndex, err.Error())
		}

		callArgs = append(callArgs, val)
	}
	return callArgs, nil
}

func (t *anyCommandTrie) parseLastArg(cmd *command, argWords []string, opt ExecuteOption, callArgs []reflect.Value) ([]reflect.Value, error) {
	if len(cmd.staticArgTypes) > 0 {
		lastArgType := cmd.staticArgTypes[len(cmd.staticArgTypes)-1]
		lastArgWords := argWords[len(cmd.staticArgTypes)-1:]
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
					return nil, fmt.Errorf("cannot parse as array element, err: %s", err.Error())
				}
				lastArg.Index(i).Set(eleVal)
			}

			if opt.ReplaceEmptyArrayWithNil && lastArg.Len() == 0 {
				lastArg = reflect.Zero(lastArgType)
			}

		} else {
			if len(lastArgWords) == 0 {
				return nil, fmt.Errorf("no enough arguments for command")
			}

			var err error
			lastArg, err = t.parseValue(lastArgWords[0], lastArgType)
			if err != nil {
				return nil, fmt.Errorf("cannot parse function argument at %d, err: %s", cmd.fnType.NumIn()-1, err.Error())
			}
		}

		callArgs = append(callArgs, lastArg)
	}
	return callArgs, nil
}

func (t *anyCommandTrie) parseValue(word string, valueType reflect.Type) (reflect.Value, error) {
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

type CommandTrie[TCtx any, TRet any] struct {
	anyTrie anyCommandTrie
}

func NewCommandTrie[TCtx any, TRet any]() CommandTrie[TCtx, TRet] {
	return CommandTrie[TCtx, TRet]{
		anyTrie: newAnyCommandTrie(reflect2.TypeOf[TCtx](), reflect2.TypeOf[TRet]()),
	}
}

func (t *CommandTrie[TCtx, TRet]) Add(fn any, prefixWords ...string) error {
	return t.anyTrie.Add(fn, prefixWords...)
}

func (t *CommandTrie[TCtx, TRet]) MustAdd(fn any, prefixWords ...string) {
	err := t.anyTrie.Add(fn, prefixWords...)
	if err != nil {
		panic(err.Error())
	}
}

func (t *CommandTrie[TCtx, TRet]) Execute(ctx TCtx, cmdWords []string, opts ...ExecuteOption) (TRet, error) {
	opt := ExecuteOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	retValues, err := t.anyTrie.Execute(ctx, cmdWords, opt)
	if err != nil {
		var defRet TRet
		return defRet, err
	}

	if retValues[0].Kind() == reflect.Interface && retValues[0].IsNil() {
		var ret TRet
		return ret, nil
	}

	return retValues[0].Interface().(TRet), nil
}

type VoidCommandTrie[TCtx any] struct {
	anyTrie anyCommandTrie
}

func NewVoidCommandTrie[TCtx any]() VoidCommandTrie[TCtx] {
	return VoidCommandTrie[TCtx]{
		anyTrie: newAnyCommandTrie(reflect2.TypeOf[TCtx](), nil),
	}
}

func (t *VoidCommandTrie[TCtx]) Add(fn any, prefixWords ...string) error {
	return t.anyTrie.Add(fn, prefixWords...)
}

func (t *VoidCommandTrie[TCtx]) MustAdd(fn any, prefixWords ...string) {
	err := t.anyTrie.Add(fn, prefixWords...)
	if err != nil {
		panic(err.Error())
	}
}

func (t *VoidCommandTrie[TCtx]) Execute(ctx TCtx, cmdWords []string, opts ...ExecuteOption) error {
	opt := ExecuteOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	_, err := t.anyTrie.Execute(ctx, cmdWords, opt)
	return err
}

type StaticCommandTrie[TRet any] struct {
	anyTrie anyCommandTrie
}

func NewStaticCommandTrie[TRet any]() StaticCommandTrie[TRet] {
	return StaticCommandTrie[TRet]{
		anyTrie: newAnyCommandTrie(nil, reflect2.TypeOf[TRet]()),
	}
}

func (t *StaticCommandTrie[TRet]) Add(fn any, prefixWords ...string) error {
	return t.anyTrie.Add(fn, prefixWords...)
}

func (t *StaticCommandTrie[TRet]) MustAdd(fn any, prefixWords ...string) {
	err := t.anyTrie.Add(fn, prefixWords...)
	if err != nil {
		panic(err.Error())
	}
}

func (t *StaticCommandTrie[TRet]) Execute(cmdWords []string, opts ...ExecuteOption) (TRet, error) {
	opt := ExecuteOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	retValues, err := t.anyTrie.Execute(nil, cmdWords, opt)
	if err != nil {
		var defRet TRet
		return defRet, err
	}

	if retValues[0].Kind() == reflect.Interface && retValues[0].IsNil() {
		var ret TRet
		return ret, nil
	}

	return retValues[0].Interface().(TRet), nil
}
