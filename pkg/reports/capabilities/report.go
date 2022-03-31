package capabilities

import (
	"encoding/json"

	"capabilities-tool/pkg"
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

	const reportType = "Cap_Level_1"
	return pkg.WriteJSON(data, r.Flags.BundleName, r.Flags.OutputPath, reportType)
}
