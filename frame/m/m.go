package m

import (
	"github.com/savorelle/maltose/container/mvar"
	"github.com/savorelle/maltose/util/mmeta"
)

type (
	Var  = mvar.Var   // Var is a universal variable interface, like generics.
	Meta = mmeta.Meta // Meta is alias of frequently-used type gmeta.Meta.
)
