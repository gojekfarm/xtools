package exporter_test

import (
	"fmt"

	"github.com/gojekfarm/xtools/xtel"
	"github.com/gojekfarm/xtools/xtel/exporter"
)

func ExampleNewOTLP() {
	_, err := xtel.NewProvider("eg-service", exporter.NewOTLP(exporter.WithTracesExporterInsecure))
	if err != nil {
		fmt.Printf("failed exporting traces to exporter: %s\n", err)
	}
}

func ExampleNewSTDOut() {
	_, err := xtel.NewProvider("eg-service", exporter.NewSTDOut(exporter.STDOutOptions{PrettyPrint: true}))
	if err != nil {
		fmt.Printf("failed printing traces to terminal: %s\n", err)
	}
}
