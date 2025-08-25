package types

import "fmt"

type Image struct {
	Repository string
	Tag        string
}

func (i Image) String() string {
	return fmt.Sprintf("%s:%s", i.Repository, i.Tag)
}
