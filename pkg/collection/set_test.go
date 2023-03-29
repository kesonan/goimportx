package collection

import (
	"strings"
	"testing"
)

func TestNewArraySet(t *testing.T) {
	instance := NewArraySet[string]()
	if instance == nil {
		t.Errorf("NewArraySet() failed, got %v, want %v", instance, "not nil")
	}
}

func TestArraySet_Add(t *testing.T) {
	set := NewArraySet[string]()
	set.Add("a")
	set.Add("b")
	set.Add("a")
	list := set.List()
	if len(list) != 2 {
		t.Errorf("Add() failed, got %v, want %v", list, "2")
	}
	if strings.Join(list, ",") != "a,b" {
		t.Errorf("Add() failed, got %v, want %v", list, "[a,b]")
	}
}
