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
	}
	s.setupExtendedTags()
}

// RegisterRuleWithTranslation 注册自定义验证规则和翻译
func (s *Server) RegisterRuleWithTranslation(rule string, fn RuleFunc, errMessage map[string]string) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册验证规则
		_ = v.RegisterValidation(rule, func(fl validator.FieldLevel) bool {
			return fn(fl)
		})

		// 注册错误翻译
		if s.translator != nil {
			for lang, msg := range errMessage {
				switch lang {
				case "zh":
					registerZhTranslation(v, s.translator, rule, msg)
				case "en":
					registerEnTranslation(v, s.translator, rule, msg)
				}
			}
		}
	}
}

// 注册中文翻译
func registerZhTranslation(v *validator.Validate, trans ut.Translator, tag string, msg string) {
	registerTranslation(v, trans, tag, msg)
}

// 注册英文翻译
func registerEnTranslation(v *validator.Validate, trans ut.Translator, tag string, msg string) {
	registerTranslation(v, trans, tag, msg)
}

// 注册翻译
func registerTranslation(v *validator.Validate, trans ut.Translator, tag string, msg string) {
	_ = v.RegisterTranslation(tag, trans, func(ut ut.Translator) error {
		return ut.Add(tag, msg, true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(tag, fe.Field())
		return t
	})
}

// 扩展 struct tag 支持
func (s *Server) setupExtendedTags() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 支持"dc"标签作为字段描述
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := fld.Tag.Get("dc")
			if name == "" {
				name = fld.Tag.Get("json")
			}
			if name == "" {
				name = fld.Name
			}
			return name
		})
	}
}
