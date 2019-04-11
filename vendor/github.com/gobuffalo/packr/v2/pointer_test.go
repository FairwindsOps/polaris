package packr

import (
	"testing"

	"github.com/gobuffalo/packr/v2/file/resolver"
	"github.com/stretchr/testify/require"
)

func Test_Pointer_Find(t *testing.T) {
	r := require.New(t)

	b1 := New("b1", "")
	r.NoError(b1.AddString("foo.txt", "FOO!"))

	b2 := New("b2", "")
	b2.SetResolver("bar.txt", &Pointer{
		ForwardBox:  "b1",
		ForwardPath: "foo.txt",
	})

	s, err := b2.FindString("bar.txt")
	r.NoError(err)
	r.Equal("FOO!", s)
}

func Test_Pointer_Find_CorrectName(t *testing.T) {
	r := require.New(t)

	gk := "0b5bab905480ad8c6d0695f615dcd644"
	g := New(gk, "")
	hgr, err := resolver.NewHexGzip(map[string]string{
		"48df4e44f4202fe5f6093beee782cb10": "1f8b08000000000000ff4c8ebdaec2300c46f7fb14bed94b5606a70b3f6283a1083186c46a2225354aad56bc3d6a2304933fdbc73ac6fffd79d7dd2f07089253fb87b5006020eb97008099c4820bb68c24465dbb63b355a07f9783cd64d414697e7211058e07a1418c9aa397603c4dd151b336df4b8992a83d514a0c372ec9a3aea345af3f7e7cb07fad21e61ec6e28cd2897bde8c53af2a5a09d4f5f777000000ffffcfb8b477d3000000",
	})
	r.NoError(err)
	g.DefaultResolver = hgr

	b := New("my box", "./templates")
	b.SetResolver("index.html", Pointer{ForwardBox: gk, ForwardPath: "48df4e44f4202fe5f6093beee782cb10"})
	f, err := b.Resolve("index.html")
	r.NoError(err)
	fi, err := f.Stat()
	r.NoError(err)
	r.Equal("index.html", fi.Name())
}
