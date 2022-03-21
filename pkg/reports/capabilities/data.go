package capabilities

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"capabilities-tool/pkg"
	"capabilities-tool/pkg/models"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

type Data struct {
	AuditCapabilities []models.AuditCapabilities
	Flags             BindFlags
}

func (d *Data) fixPackageNameInconsistency() {
	for _, auditCapabilities := range d.AuditCapabilities {
		if auditCapabilities.PackageName == "" {
			split := strings.Split(auditCapabilities.OperatorBundleImagePath, "/")
			nm := ""
			for _, v := range split {
				if strings.Contains(v, "@") {
					nm = strings.Split(v, "@")[0]
					break
				}
			}
			for _, bundle := range d.AuditCapabilities {
				if strings.Contains(bundle.OperatorBundleImagePath, nm) {
					auditCapabilities.PackageName = bundle.PackageName
				}
			}
		}
	}
}

func (d *Data) PrepareReport() Report {
	d.fixPackageNameInconsistency()

	var allBundle []Bundle
	for _, v := range d.AuditCapabilities {
		col := NewBundle(v)

		allBundle = append(allBundle, *col)
	}

	sort.Slice(allBundle[:], func(i, j int) bool {
		return allBundle[i].PackageName < allBundle[j].PackageName
	})

	finalReport := Report{}
	finalReport.Flags = d.Flags
	finalReport.Bundles = allBundle

	dt := time.Now().Format("2006-01-02")
	finalReport.GenerateAt = dt

	if len(allBundle) == 0 {
		log.Fatal("No data was found for the criteria informed. " +
			"Please, ensure that you provide valid information.")
	}

	return finalReport
}

func (d *Data) OutputReport() error {
	report := d.PrepareReport()
	switch d.Flags.OutputFormat {
	case pkg.JSON:
		if err := report.writeJSON(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format : %s", d.Flags.OutputFormat)
	}
	return nil
}

func (d *Data) BuildCapabilitiesQuery() (string, error) {
	query := sq.Select("o.name, o.bundlepath").From(
		"operatorbundle o")

	if d.Flags.HeadOnly {
		query = sq.Select("o.name, o.bundlepath").From(
			"operatorbundle o, channel c").Limit(30)
		query = query.Where("c.head_operatorbundle_name == o.name")
	}
	if d.Flags.Limit > 0 {
		query = query.Limit(uint64(d.Flags.Limit))
	}
	if len(d.Flags.Filter) > 0 {
		query = sq.Select("o.name, o.bundlepath").From(
			"operatorbundle o, channel_entry c")
		like := "'%" + d.Flags.Filter + "%'"
		query = query.Where(fmt.Sprintf("c.operatorbundle_name == o.name AND c.package_name like %s", like))
	}

	if len(d.Flags.FilterBundle) > 0 {
		query = sq.Select("o.name, o.bundlepath").From(
			"operatorbundle o, channel_entry c")
		like := "'%" + d.Flags.FilterBundle + "%'"
		query = query.Where(fmt.Sprintf("c.operatorbundle_name == o.name AND c.operatorbundle_name like %s", like))
	}

	query.OrderBy("o.name")

	sql, _, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("unable to create sql : %s", err)
	}
	return sql, nil
}
