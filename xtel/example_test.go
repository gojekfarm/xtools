package xtel_test

import "github.com/gojekfarm/xtools/xtel"

func ExampleNewProvider() {
	tp, err := xtel.NewProvider(
		"service-a",
		xtel.DisableClientAutoTracing,
		xtel.SamplingFraction(0.1),
	)
	if err != nil {
		panic(err)
	}

	if err := tp.Start(); err != nil {
		panic(err)
	}

	defer tp.Stop()
}
