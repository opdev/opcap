package capabilities

import (
	"capabilities-tool/pkg/models"
)

type Bundle struct {
	BundleName      string   `json:"bundleName,omitempty"`
	PackageName     string   `json:"packageName,omitempty"`
	BundleImagePath string   `json:"bundleImagePath,omitempty"`
	DefaultChannel  string   `json:"defaultChannel,omitempty"`
	Channels        []string `json:"bundleChannel,omitempty"`
	AuditErrors     []string `json:"errors,omitempty"`
	Capabilities    bool     `json:"OperatorInstalled"`
	InstallLogs     []string `json:"InstallLogs"`
	CleanUpLogs     []string `json:"CleanUpLogs,omitempty"`
}

func NewBundle(v models.AuditCapabilities) *Bundle {
	col := Bundle{}
	col.BundleName = v.OperatorBundleName
	col.PackageName = v.PackageName
	col.BundleImagePath = v.OperatorBundleImagePath
	//col.DefaultChannel = v.DefaultChannel
	//col.Channels = v.Channels
	//col.AuditErrors = v.Errors
	col.Capabilities = v.Capabilities
	col.InstallLogs = v.InstallLogs
	col.CleanUpLogs = v.CleanUpLogs

	return &col
}
