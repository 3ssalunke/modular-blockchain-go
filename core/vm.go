package core

import (
	"fmt"

	"github.com/3ssalunke/go-blockchain/util"
)

type Instruction byte

const (
	InstrPushInt  Instruction = 0x0a
	InstrPushByte Instruction = 0x0b

	InstrPack Instruction = 0x10

	InstrStore Instruction = 0x20
	InstrGet   Instruction = 0x21

	InstrAdd Instruction = 0x30
	InstrSub Instruction = 0x32
	InstrMul Instruction = 0x33
	InstrDiv Instruction = 0x34
)

type Stack struct {
	data []any
	sp   int
}

func NewStack(size int) *Stack {
	return &Stack{
		data: make([]any, size),
		sp:   0,
	}
}

func (s *Stack) Push(v any) {
	// s.data[s.sp] = v
	s.data = append([]any{v}, s.data...)
	s.sp++
}

func (s *Stack) Pop() any {
	value := s.data[0]
	s.data = append(s.data[:0], s.data[1:]...)
	s.sp--
	return value
}

type VM struct {
	data          []byte
	ip            int
	stack         *Stack
	contractState *State
}

func NewVM(data []byte, contractState *State) *VM {
	return &VM{
		data:          data,
		ip:            0,
		stack:         NewStack(128),
		contractState: contractState,
	}
}

func (vm *VM) Run() error {
	for {
		instr := vm.data[vm.ip]

		if err := vm.Exec(Instruction(instr)); err != nil {
			return err
		}

		vm.ip++

		if vm.ip > len(vm.data)-1 {
			break
		}
	}

	return nil
}

func (vm *VM) Exec(instr Instruction) error {
	fmt.Println(instr)

	switch instr {
	case InstrStore:
		var (
			key             = vm.stack.Pop().([]byte)
			value           = vm.stack.Pop()
			serializedValue []byte
		)
		switch v := value.(type) {
		case int:
			serializedValue = util.SerializeInt64(int64(v))
		default:
			panic("TODO: unknown type")
		}
		vm.contractState.Put(key, serializedValue)
	case InstrPushInt:
		vm.stack.Push(int(vm.data[vm.ip-1]))
	case InstrPushByte:
		vm.stack.Push(byte(vm.data[vm.ip-1]))
	case InstrPack:
		n := vm.stack.Pop().(int)
		b := make([]byte, n)

		for i := 0; i < n; i++ {
			b[i] = vm.stack.Pop().(byte)
		}

		vm.stack.Push(b)
	case InstrAdd:
		a := vm.stack.Pop().(int) + vm.stack.Pop().(int)
		vm.stack.Push(a)
	case InstrSub:
		c := vm.stack.Pop().(int)
		d := vm.stack.Pop().(int)
		a := c - d
		vm.stack.Push(a)
	case InstrMul:
		a := vm.stack.Pop().(int) * vm.stack.Pop().(int)
		vm.stack.Push(a)
	case InstrDiv:
		c := vm.stack.Pop().(int)
		d := vm.stack.Pop().(int)
		a := c / d
		vm.stack.Push(a)
	}

	return nil
}
