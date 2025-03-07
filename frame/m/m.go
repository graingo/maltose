package m

import (
	"github.com/graingo/maltose/container/mvar"
	"github.com/graingo/maltose/util/mmeta"
)

type (
	Var  = mvar.Var   // Var is a universal variable interface, like generics.
	Meta = mmeta.Meta // Meta is alias of frequently-used type gmeta.Meta.
)
