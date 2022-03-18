package capabilities

import (
	"encoding/json"

	"awesomeProject/pkg"
)

type Report struct {
	Bundles    []Bundle
	Flags      BindFlags
	GenerateAt string
}

func (r *Report) writeJSON() error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}

	const reportType = "capabilities"
	return pkg.WriteJSON(data, r.Flags.FilterBundle, r.Flags.OutputPath, reportType)
}
