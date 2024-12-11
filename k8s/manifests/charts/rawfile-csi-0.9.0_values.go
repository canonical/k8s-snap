package main

type RawfileCsi090Values_RawfileCsi090Values_Defaults_Image struct {
	Repository any
	Tag        any
	PullPolicy any
}

type RawfileCsi090Values_RawfileCsi090Values_Defaults_RawfileCsi090Values_RawfileCsi090Values_Defaults_Resources_Limits struct {
	Cpu    any
	Memory any
}

type RawfileCsi090Values_RawfileCsi090Values_Defaults_RawfileCsi090Values_RawfileCsi090Values_Defaults_Resources_Requests struct {
	Cpu    any
	Memory any
}

type RawfileCsi090Values_RawfileCsi090Values_Defaults_Resources struct {
	Limits   RawfileCsi090Values_RawfileCsi090Values_Defaults_RawfileCsi090Values_RawfileCsi090Values_Defaults_Resources_Limits
	Requests RawfileCsi090Values_RawfileCsi090Values_Defaults_RawfileCsi090Values_RawfileCsi090Values_Defaults_Resources_Requests
}

type RawfileCsi090Values_Defaults struct {
	Image     RawfileCsi090Values_RawfileCsi090Values_Defaults_Image
	Resources RawfileCsi090Values_RawfileCsi090Values_Defaults_Resources
}

type RawfileCsi090Values_RawfileCsi090Values_Controller_CsiDriverArgsItem struct {
}

type RawfileCsi090Values_Controller struct {
	any
	CsiDriverArgs []RawfileCsi090Values_RawfileCsi090Values_Controller_CsiDriverArgsItem
}

type RawfileCsi090Values_Images struct {
	CsiNodeDriverRegistrar any
	CsiProvisioner         any
	CsiResizer             any
	CsiSnapshotter         any
}

type RawfileCsi090Values_RawfileCsi090Values_Node_Storage struct {
	Path any
}

type RawfileCsi090Values_RawfileCsi090Values_Node_Metrics struct {
	Enabled any
}

type RawfileCsi090Values_Node struct {
	any
	Storage RawfileCsi090Values_RawfileCsi090Values_Node_Storage
	Metrics RawfileCsi090Values_RawfileCsi090Values_Node_Metrics
}

type RawfileCsi090Values_StorageClass struct {
	Enabled           any
	Name              any
	IsDefault         any
	ReclaimPolicy     any
	VolumeBindingMode any
}

type RawfileCsi090Values_ServiceMonitor struct {
	Enabled  any
	Interval any
}

type RawfileCsi090Values struct {
	ProvisionerName  any
	Defaults         RawfileCsi090Values_Defaults
	Controller       RawfileCsi090Values_Controller
	Images           RawfileCsi090Values_Images
	Node             RawfileCsi090Values_Node
	StorageClass     RawfileCsi090Values_StorageClass
	ImagePullSecrets []any
	ServiceMonitor   RawfileCsi090Values_ServiceMonitor
}
