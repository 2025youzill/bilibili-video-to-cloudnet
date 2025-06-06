package spew

import (
	"bvtc/config"

	"github.com/davecgh/go-spew/spew"
)

func InitSpew(){
	cfg := config.GetConfig()
	spew.Config.Indent = cfg.Spew.Indent
	spew.Config.MaxDepth = cfg.Spew.MaxDepth
	spew.Config.DisableMethods = cfg.Spew.DisableMethods
	spew.Config.DisablePointerMethods = cfg.Spew.DisablePointerMethods
	spew.Config.ContinueOnMethod = cfg.Spew.ContinueOnMethod
	spew.Config.SortKeys = cfg.Spew.SortKeys 
}