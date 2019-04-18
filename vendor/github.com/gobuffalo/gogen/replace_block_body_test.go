package gogen

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/gobuffalo/genny"
	"github.com/stretchr/testify/require"
)

func Test_ReplaceBlockBody(t *testing.T) {
	r := require.New(t)

	path := "actions/app.go"
	f := genny.NewFile(path, strings.NewReader(modelBeforeBodyReplace))

	f, err := ReplaceBlockBody(f, "func (u *User) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {", `
var err error
return validate.Validate(
	&validators.StringIsPresent{Field: u.Password, Name: "Password"},
	&validators.StringsMatch{Name: "Password", Field: u.Password, Field2: u.PasswordConfirmation, Message: "Password does not match confirmation"},
), err`)

	r.NoError(err)

	b, err := ioutil.ReadAll(f)
	r.NoError(err)

	r.Equal(path, f.Name())
	r.Equal(modelAfterBodyReplace, string(b))
}

const modelBeforeBodyReplace = `
package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type User struct {
	ID           uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Email        string
	PasswordHash string
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (u *User) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}`

const modelAfterBodyReplace = `
package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type User struct {
	ID           uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Email        string
	PasswordHash string
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (u *User) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {

var err error
return validate.Validate(
	&validators.StringIsPresent{Field: u.Password, Name: "Password"},
	&validators.StringsMatch{Name: "Password", Field: u.Password, Field2: u.PasswordConfirmation, Message: "Password does not match confirmation"},
), err
}`
