package mhttp

import (
	"reflect"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

// RuleFunc 自定义验证规则函数
type RuleFunc func(fl validator.FieldLevel) bool

// registerValidateTranslator 注册 gin validator 翻译器
func (s *Server) registerValidateTranslator(locale string) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		zhT := zh.New()
		enT := en.New()
		uni := ut.New(enT, zhT, enT)
		trans, _ := uni.GetTranslator(locale)
		s.translator = trans
		switch locale {
		case "en":
			_ = en_translations.RegisterDefaultTranslations(v, trans)
		case "zh":
			_ = zh_translations.RegisterDefaultTranslations(v, trans)
		}
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			return fld.Tag.Get("dc")
		})
	}
}

// RegisterRule 注册自定义验证规则
func (s *Server) RegisterRule(rule string, fn RuleFunc) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation(rule, func(fl validator.FieldLevel) bool {
			return fn(fl)
		})
	}
}
