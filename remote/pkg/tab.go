package pkg

import (
	"github.com/neovim/go-client/nvim"
)

func CreateNewTab(v *nvim.Nvim, buf nvim.Buffer) error {
	if err := v.SetCurrentBuffer(buf); err != nil {
		return err
	}

	return nil
}
