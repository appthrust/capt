package v1beta2

// WorkspaceTemplateReference contains the reference to WorkspaceTemplate
type WorkspaceTemplateReference struct {
	// Name of the referenced WorkspaceTemplate
	Name string `json:"name"`
	// Namespace of the referenced WorkspaceTemplate
	Namespace string `json:"namespace,omitempty"`
}

// WorkspaceReference defines a reference to a Workspace
type WorkspaceReference struct {
	// Name of the referenced Workspace
	Name string `json:"name"`
	// Namespace of the referenced Workspace
	Namespace string `json:"namespace,omitempty"`
}
