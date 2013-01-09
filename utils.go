package goweb

import (
	"encoding/json"
	"strings"
)

type Module interface {}

func isRestNew(id string)bool{
	return id == "new"
}

func isRestEdit(id string)bool{
	return strings.Contains(id,";edit")
}

