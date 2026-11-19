package vsnip

import (
	"context"
	"reflect"
	"testing"

	"github.com/Shougo/goda-cmp.vim/cmp"
	"github.com/Shougo/goda-cmp.vim/cmp/cmptest"
)

// vsnipGroups mimics what h.Call decodes from vsnip#source#find() over the
// wire: a list of lists of dicts, generic maps and slices rather than the
// package's own snippet struct, to exercise sdk.Decode's json round trip.
func vsnipGroups() []any {
	return []any{
		[]any{
			map[string]any{
				"label":       "today",
				"prefix":      []any{"today"},
				"body":        []any{"2026-11-19"},
				"description": "",
			},
			map[string]any{
				"label":       "Class",
				"prefix":      []any{"class"},
				"body":        []any{"class $1 {", "  $0", "}"},
				"description": "Class definition template.",
			},
		},
		[]any{
			// A duplicate prefix from a second source: the first one found
			// (above) must win.
			map[string]any{
				"label":       "today (dup)",
				"prefix":      []any{"today"},
				"body":        []any{"DUPLICATE"},
				"description": "",
			},
		},
	}
}

func TestGather(t *testing.T) {
	// cmptest.Host hands args to the fake as the Go value the source
	// passed to h.Call, with no JSON round trip (that only happens on the
	// Decode side): cmp.Context.Bufnr is a plain int, so this is too.
	var gotBufnr int
	h := &cmptest.Host{Funcs: map[string]func(...any) (any, error){
		"vsnip#source#find": func(a ...any) (any, error) {
			gotBufnr, _ = a[0].(int)
			return vsnipGroups(), nil
		},
	}}
	g, err := New().Gather(context.Background(), h, &cmp.Context{Bufnr: 7, Input: "to"})
	if err != nil {
		t.Fatal(err)
	}
	if gotBufnr != 7 {
		t.Errorf("vsnip#source#find bufnr = %d, want 7", gotBufnr)
	}
	if g.Start != 0 {
		t.Errorf("Start = %d, want 0", g.Start)
	}
	var words, menus, kinds []string
	var dups []int
	for _, it := range g.Items {
		words = append(words, it.Word)
		menus = append(menus, it.Menu)
		kinds = append(kinds, it.Kind)
		dups = append(dups, it.Dup)
	}
	wantWords := []string{"today", "class"}
	if !reflect.DeepEqual(words, wantWords) {
		t.Errorf("words = %q, want %q (duplicate prefix must be dropped)", words, wantWords)
	}
	if !reflect.DeepEqual(menus, []string{"[v] today", "[v] Class definition template."}) {
		t.Errorf("menus = %q", menus)
	}
	for _, k := range kinds {
		if k != "Snippet" {
			t.Errorf("kind = %q, want Snippet", k)
		}
	}
	for _, d := range dups {
		if d != 1 {
			t.Errorf("dup = %d, want 1", d)
		}
	}
	if body, _ := g.Items[1].UserData.(string); body != "class $1 {\n  $0\n}" {
		t.Errorf("UserData = %q", body)
	}
}

func TestOnCompleteDone(t *testing.T) {
	var gotBody string
	var gotOpts map[string]any
	h := &cmptest.Host{Funcs: map[string]func(...any) (any, error){
		"vsnip#anonymous": func(a ...any) (any, error) {
			gotBody, _ = a[0].(string)
			gotOpts, _ = a[1].(map[string]any)
			return nil, nil
		},
	}}
	item := cmp.Item{Word: "class", UserData: "class $1 {\n  $0\n}"}
	if err := New().(cmp.CompleteDoner).OnCompleteDone(context.Background(), h, cmp.DoneArgs{Item: item}); err != nil {
		t.Fatal(err)
	}
	if gotBody != "class $1 {\n  $0\n}" {
		t.Errorf("body = %q", gotBody)
	}
	if gotOpts["prefix"] != "class" {
		t.Errorf("opts = %+v", gotOpts)
	}
}

func TestOnCompleteDoneNoBody(t *testing.T) {
	h := &cmptest.Host{Funcs: map[string]func(...any) (any, error){}}
	item := cmp.Item{Word: "foo"} // UserData from another source, or none
	if err := New().(cmp.CompleteDoner).OnCompleteDone(context.Background(), h, cmp.DoneArgs{Item: item}); err != nil {
		t.Fatal(err)
	}
	if len(h.Calls()) != 0 {
		t.Errorf("calls = %+v, want none", h.Calls())
	}
}
