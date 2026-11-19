// Package vsnip is a goda-cmp source for the snippets hrsh7th/vim-vsnip
// finds for the buffer (vsnip#source#find): a port of the completion items
// vsnip#get_complete_items() builds for other completion engines (ddc.vim,
// asyncomplete.vim, ...). Kind is vsnip's own "Snippet"; unlike vsnip's own
// convention, Menu is left for SourceOption.Mark to set (as most goda-cmp-
// ext sources do) and the description goes to Info instead, so a
// configuration controls the inline tag itself rather than it being baked
// in here.
//
// Accepting a candidate removes the prefix goda-cmp just inserted and
// expands the snippet with vsnip#anonymous({body}, {'prefix': word}),
// exactly as vsnip's own <Plug>(vsnip-expand) removes the typed prefix
// before expanding. vim-vsnip itself must be on 'runtimepath' (it has no Go
// side); this package only calls its autoload functions.
package vsnip

import (
	"context"
	"strings"

	"github.com/Shougo/goda-cmp.vim/cmp"
	"github.com/Shougo/goda.vim/sdk"
)

// snippet is one entry of vsnip#source#find()'s result: a list of lists of
// these dicts (one inner list per source: user_snippet, vscode, snipmate).
// prefix_alias is not read, matching vsnip#get_complete_items() itself,
// which only ever offers .prefix as a candidate.
type snippet struct {
	Label       string   `json:"label"`
	Prefix      []string `json:"prefix"`
	Body        []string `json:"body"`
	Description string   `json:"description"`
}

type source struct{}

// New returns the source, named "vsnip".
func New() cmp.Source { return source{} }

func (source) Name() string { return "vsnip" }

func (source) Gather(ctx context.Context, h sdk.Host, c *cmp.Context) (cmp.Gathered, error) {
	start := c.KeywordStart()

	v, err := h.Call(ctx, "vsnip#source#find", c.Bufnr)
	if err != nil {
		return cmp.Gathered{Start: start}, err
	}
	var groups [][]snippet
	if err := sdk.Decode(v, &groups); err != nil {
		return cmp.Gathered{Start: start}, err
	}

	// vsnip#get_complete_items() dedupes by prefix across every snippet of
	// every source the same way: the first one found keeps it.
	seen := make(map[string]bool)
	var items []cmp.Item
	for _, group := range groups {
		for _, sn := range group {
			info := sn.Description
			if info == "" {
				info = sn.Label
			}
			body := strings.Join(sn.Body, "\n")
			for _, prefix := range sn.Prefix {
				if seen[prefix] {
					continue
				}
				seen[prefix] = true
				items = append(items, cmp.Item{
					Word:     prefix,
					Kind:     "Snippet",
					Info:     info,
					Dup:      1,
					UserData: body,
				})
			}
		}
	}
	return cmp.Gathered{Start: start, Items: items}, nil
}

// OnCompleteDone removes the inserted prefix and expands the snippet's
// body in its place, the same two steps vsnip#expand() itself does for a
// typed-and-triggered snippet.
func (source) OnCompleteDone(ctx context.Context, h sdk.Host, a cmp.DoneArgs) error {
	body, ok := a.Item.UserData.(string)
	if !ok || body == "" {
		return nil
	}
	_, err := h.Call(ctx, "vsnip#anonymous", body, map[string]any{"prefix": a.Item.Word})
	return err
}

var (
	_ cmp.Source        = source{}
	_ cmp.CompleteDoner = source{}
)
