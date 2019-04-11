package import_pkg

import (
	"github.com/gobuffalo/packr/v2"
)

var BoxTestNew = packr.New("pkg_test", "./pkg_test")
var BoxTestNewBox = packr.NewBox("./pkg_test")
