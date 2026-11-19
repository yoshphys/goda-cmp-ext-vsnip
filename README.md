# goda-cmp-ext-vsnip.vim

> [!NOTE]
> goda-cmp.vim itself is under active development, so the API of this plugin
> may change to follow it.

A [goda-cmp.vim](https://github.com/Shougo/goda-cmp.vim) source for the
snippets [hrsh7th/vim-vsnip](https://github.com/hrsh7th/vim-vsnip) finds for
the buffer: a port of the completion items `vsnip#get_complete_items()`
builds for other completion engines (ddc.vim, asyncomplete.vim, ...), using
vsnip's own convention (`kind` "Snippet", `menu` `"[v] {description or
label}"`).

Accepting a candidate removes the prefix goda-cmp just inserted and expands
the snippet with `vsnip#anonymous({body}, {'prefix': word})`, exactly as
vsnip's own `<Plug>(vsnip-expand)` removes the typed prefix before expanding.

This plugin is pure Go: vim-vsnip itself must be on `'runtimepath'` already
(install it with your plugin manager as usual); nothing here needs installing
as a Vim plugin.

## Requirements

- goda.vim and goda-cmp.vim, and Go to build your configuration
- [hrsh7th/vim-vsnip](https://github.com/hrsh7th/vim-vsnip) on `'runtimepath'`

## Installation

```sh
go get github.com/yoshphys/goda-cmp-ext-vsnip.vim@latest
```

## Configuration

```go
import (
	"github.com/Shougo/goda-cmp.vim/cmp"
	"github.com/yoshphys/goda-cmp-ext-vsnip.vim/ext/source/vsnip"
)

func main() {
	b := cmp.New()
	// ...
	b.AddSource(vsnip.New())
	b.PatchGlobal(cmp.Config{
		Sources: []string{"vsnip" /* , ... */},
		SourceOptions: map[string]cmp.SourceOption{
			"vsnip": {Mark: "[v]"},
		},
	})
	// ...
}
```

## License

MIT. See [LICENSE](LICENSE).
