package tiptop

import (
	"fmt"
	"github.com/niubaoshu/gotiny"
	"testing"
)

func TestNewTipTop(t *testing.T) {
	tip, err := NewTipTop(DefaultConfig())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(tip)
}

func TestTipTop_Get(t *testing.T) {
	tip, _ := NewTipTop(DefaultConfig())
	insertData(tip)
	out := make(map[string]string)
	bytes, err := tip.Get("key1")
	gotiny.Unmarshal(bytes, &out)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("输出：%v\n", out)

	fmt.Printf("Stats:%v", tip.GetStats())
}

func TestTipTop_Delete(t *testing.T) {
	tip, err := NewTipTop(DefaultConfig())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	insertData(tip)
	out := make(map[string]string)
	err = tip.Delete("key")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	out = make(map[string]string)
	bytes, err := tip.Get("key")
	gotiny.Unmarshal(bytes, &out)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("输出：%v", out)
}

func TestTipTop_Reset(t *testing.T) {
	tip, err := NewTipTop(DefaultConfig())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	insertData(tip)
	tip.Reset()

	out := make(map[string]string)
	bytes, err := tip.Get("key1")

	if err != nil {
		fmt.Println(err.Error())
	} else {
		gotiny.Unmarshal(bytes, &out)
	}
	fmt.Printf("输出：%v\n", out)

	bytes, err = tip.Get("key2")

	if err != nil {
		fmt.Println(err.Error())
	} else {
		gotiny.Unmarshal(bytes, &out)
	}
	fmt.Printf("输出：%v", out)
}

func TestTipTop_Len(t *testing.T) {
	tip, err := NewTipTop(DefaultConfig())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	insertData(tip)

	fmt.Printf("容量：%v", tip.Len())
}

func TestTipTop_Cap(t *testing.T) {
	tip, err := NewTipTop(DefaultConfig())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	insertData(tip)

	fmt.Printf("容量：%v", tip.Cap())
}

func insertData(t *TipTop) {
	value := map[string]string{
		"key1": "value1",
		"key2": "value12",
	}
	bytes := gotiny.Marshal(&value)
	_ = t.Set("key1", bytes)
	_ = t.Set("key2", bytes)
}
