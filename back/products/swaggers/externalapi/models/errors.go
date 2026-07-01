package models

import "go/types"

type Err400 struct {
	// @tg example=``
	Data types.Nil `json:"data,omitempty"`
	// @tg desc=`Флаг показывающий, что ответ пришел с ошибкой`
	// @tg example=`true`
	Error bool `json:"error"`
	// @tg desc=`Заголовок ошибки`
	// @tg example=`content.api.errors.products.badRequest`
	ErrorText string `json:"errorText"`
	// @tg desc=`Текст ошибки, при ответе`
	AdditionalErrors struct {
		Errors []struct {
			TrKey string `json:"trKey"`
			// @tg example=`{"1": "value one", "2": "value two"}`
			Params map[string]string `json:"params"`
		} `json:"errors"`
	} `json:"additionalErrors"`
}

type Err401 struct {
	// @tg example=``
	Data types.Nil `json:"data,omitempty"`
	// @tg desc=`Флаг показывающий, что ответ пришел с ошибкой`
	// @tg example=`true`
	Error bool `json:"error"`
	// @tg desc=`Заголовок ошибки`
	// @tg example=`content.api.errors.products.unauthorized`
	ErrorText string `json:"errorText"`
	// @tg desc=`Текст ошибки, при ответе, со статус кодом 401, не указывается`
	AdditionalErrors types.Nil `json:"additionalErrors,omitempty"`
}

type Err403 struct {
	// @tg example=``
	Data types.Nil `json:"data,omitempty"`
	// @tg desc=`Флаг показывающий, что ответ пришел с ошибкой`
	// @tg example=`true`
	Error bool `json:"error"`
	// @tg desc=`Заголовок ошибки`
	// @tg example=`content.api.errors.products.accessDenied`
	ErrorText string `json:"errorText"`
	// @tg desc=`Текст ошибки, при ответе, со статус кодом 403, не указывается`
	AdditionalErrors types.Nil `json:"additionalErrors,omitempty"`
}

type Err500 struct {
	// @tg example=``
	Data types.Nil `json:"data,omitempty"`
	// @tg desc=`Флаг показывающий, что ответ пришел с ошибкой`
	// @tg example=`true`
	Error bool `json:"error"`
	// @tg desc=`Заголовок ошибки`
	// @tg example=`content.api.errors.products.internalError`
	ErrorText string `json:"errorText"`
	// @tg desc=`Текст ошибки, при ответе, со статус кодом 500, не указывается`
	AdditionalErrors types.Nil `json:"additionalErrors,omitempty"`
}
