package models

import (
	"go/types"
)

type Resp200 struct {
	// @tg desc=`Флаг показывающий, что ответ пришел с ошибкой`
	// @tg example=false
	Error bool `json:"error"`
	// @tg example=``
	ErrorText string `json:"errorText"`
	// @tg example=`true`
	Data types.Nil `json:"data"`
	// @tg example=``
	AdditionalErrors types.Nil `json:"additionalErrors"`
}
