package handlers

import "go.uber.org/fx"

var Module = fx.Module(
	"handlers",
	fx.Provide(NewPayments),
	fx.Provide(NewUsers),
	fx.Provide(NewHome),
)
