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

// registerValidateTranslator registers the gin validator translator, making sure it only runs once.
func (s *Server) registerValidateTranslator(locale string) {
	if s.uni != nil {
		// If uni is already initialized, just ensure the server's default translator is set.
		if trans, found := s.uni.GetTranslator(locale); found {
			s.translator = trans
		}
		return
	}
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		zhT := zh.New()
		enT := en.New()
		s.uni = ut.New(enT, zhT, enT)
		trans, _ := s.uni.GetTranslator(locale)
		s.translator = trans

		// Register default translations for all supported languages
		_ = en_translations.RegisterDefaultTranslations(v, s.uni.GetFallback())
		if zhTrans, found := s.uni.GetTranslator("zh"); found {
			_ = zh_translations.RegisterDefaultTranslations(v, zhTrans)
		}
	}
	s.setupExtendedTags()
}

// RegisterRuleWithTranslation registers the custom validation rule and translation for multiple languages.
func (s *Server) RegisterRuleWithTranslation(rule string, fn RuleFunc, errMessage map[string]string) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Register the validation rule function itself.
		_ = v.RegisterValidation(rule, func(fl validator.FieldLevel) bool {
			return fn(fl)
		})

		// Ensure the universal translator is initialized.
		if s.uni == nil {
			s.registerValidateTranslator(s.config.ServerLocale)
		}

		// Register translations for each language provided.
		for lang, msg := range errMessage {
			if trans, found := s.uni.GetTranslator(lang); found {
				registerTranslation(v, trans, rule, msg)
			}
		}
	}
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
