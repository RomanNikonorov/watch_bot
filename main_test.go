package main

import (
	"reflect"
	"testing"
)

func TestParseSemicolonSeparatedList(t *testing.T) {
	result := parseSemicolonSeparatedList(" user-1;user-2 ; ; user-3 ")
	expected := []string{"user-1", "user-2", "user-3"}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestParseSemicolonSeparatedListDoesNotSplitCommas(t *testing.T) {
	result := parseSemicolonSeparatedList("user-1,user-2")
	expected := []string{"user-1,user-2"}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}
