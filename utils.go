package goweb

import (
	"strings"
)

type Module interface{}

func isRestNew(id string) bool {
	return id == "new"
}

func isRestEdit(id string) bool {
	return strings.Contains(id, ";edit")
}
