package capabilities

// BindFlags define the flags used to generate the bundle report
type BindFlags struct {
	//IndexImage      string `json:"image"`
	//Limit           int32  `json:"limit"`
	//HeadOnly        bool   `json:"headOnly"`
	//Filter          string `json:"filter"`
	//Namespace       string `json:"namespace"`
	BundleName      string `json:"bundleName,omitempty"`
	Endpoint        string `json:"endpoint"`
	S3Bucket        string `json:"s3Bucket"`
	InstallMode     string `json:"installMode"`
	FilterBundle    string `json:"FilterBundle"`
	OutputPath      string `json:"outputPath"`
	OutputFormat    string `json:"outputFormat"`
	ContainerEngine string `json:"containerEngine"`
	PullSecretName  string `json:"pullSecretName"`
	ServiceAccount  string `json:"serviceAccount"`
	PackageName     string `json:"packageName,omitempty"`
}
