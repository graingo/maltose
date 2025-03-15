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

// RuleFunc is the custom validation rule function.
type RuleFunc func(fl validator.FieldLevel) bool

// registerValidateTranslator registers the gin validator translator.
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

// RegisterRuleWithTranslation registers the custom validation rule and translation.
func (s *Server) RegisterRuleWithTranslation(rule string, fn RuleFunc, errMessage map[string]string) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// register validation rule
		_ = v.RegisterValidation(rule, func(fl validator.FieldLevel) bool {
			return fn(fl)
		})

		// register error translation
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

// registerZhTranslation registers the Chinese translation.
func registerZhTranslation(v *validator.Validate, trans ut.Translator, tag string, msg string) {
	registerTranslation(v, trans, tag, msg)
}

// registerEnTranslation registers the English translation.
func registerEnTranslation(v *validator.Validate, trans ut.Translator, tag string, msg string) {
	registerTranslation(v, trans, tag, msg)
}

// registerTranslation registers the translation.
func registerTranslation(v *validator.Validate, trans ut.Translator, tag string, msg string) {
	_ = v.RegisterTranslation(tag, trans, func(ut ut.Translator) error {
		return ut.Add(tag, msg, true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(tag, fe.Field())
		return t
	})
}

// setupExtendedTags extends the struct tag support.
func (s *Server) setupExtendedTags() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// support "dc" tag as field description
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
