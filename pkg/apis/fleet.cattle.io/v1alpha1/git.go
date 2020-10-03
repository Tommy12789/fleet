package v1alpha1

import (
	"github.com/rancher/wrangler/pkg/genericcondition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	RepoLabel            = "fleet.cattle.io/repo-name"
	BundleNamespaceLabel = "fleet.cattle.io/bundle-namespace"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type GitRepo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitRepoSpec   `json:"spec,omitempty"`
	Status GitRepoStatus `json:"status,omitempty"`
}

type GitRepoSpec struct {
	// Repo is a URL to a git repo to clone and index
	Repo string `json:"repo,omitempty"`

	// Branch The git branch to follow
	Branch string `json:"branch,omitempty"`

	// Revision A specific commit or tag to operate on
	Revision string `json:"revision,omitempty"`

	// Ensure that all resources are created in this namespace
	// Any cluster scoped resource will be rejected if this is set
	// Additionally this namespace will be created on demand
	TargetNamespace string `json:"targetNamespace,omitempty"`

	// ClientSecretName is the client secret to be used to connect to the repo
	// It is expected the secret be of type "kubernetes.io/basic-auth" or "kubernetes.io/ssh-auth".
	ClientSecretName string `json:"clientSecretName,omitempty"`

	// Paths is the directories relative to the git repo root that contain resources to be applied.
	// Path globbing is support, for example ["charts/*"] will match all folders as a subdirectory of charts/
	// If empty, "/" is the default
	Paths []string `json:"paths,omitempty"`

	// Paused this cause changes in Git to not be propagated down to the clusters but instead mark
	// resources as OutOfSync
	Paused bool `json:"paused,omitempty"`

	// ServiceAccount used in the downstream cluster for deployment
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// Targets is a list of target this repo will deploy to
	Targets []GitTarget `json:"targets,omitempty"`

	// PollingInterval is how often to check git for new updates
	PollingInterval *metav1.Duration `json:"pollingInterval,omitempty"`

	// All non-ready deployments before this time will be resynced
	ForceSyncBefore *metav1.Time `json:"forceSyncBefore,omitempty"`

	// ForceUpdate is a timestamp if set to Now() will cause the git repo to be checked again
	ForceUpdate *metav1.Time `json:"forceUpdate,omitempty"`
}

type GitTarget struct {
	Name                 string                `json:"name,omitempty"`
	ClusterSelector      *metav1.LabelSelector `json:"clusterSelector,omitempty"`
	ClusterGroup         string                `json:"clusterGroup,omitempty"`
	ClusterGroupSelector *metav1.LabelSelector `json:"clusterGroupSelector,omitempty"`
}

type GitRepoStatus struct {
	ObservedGeneration   int64                               `json:"observedGeneration"`
	Commit               string                              `json:"commit,omitempty"`
	ReadyClusters        int                                 `json:"readyClusters"`
	DesiredReadyClusters int                                 `json:"desiredReadyClusters"`
	Summary              BundleSummary                       `json:"summary,omitempty"`
	Display              GitRepoDisplay                      `json:"display,omitempty"`
	Conditions           []genericcondition.GenericCondition `json:"conditions,omitempty"`
	Resources            []GitRepoResource                   `json:"resources,omitempty"`
	ResourceCounts       GitRepoResourceCounts               `json:"resourceCounts,omitempty"`
	ResourceErrors       []string                            `json:"resourceErrors,omitempty"`
}

type GitRepoResourceCounts struct {
	Ready        int `json:"ready"`
	DesiredReady int `json:"desiredReady"`
	WaitApplied  int `json:"waitApplied"`
	Modified     int `json:"modified"`
	Orphaned     int `json:"orphaned"`
	Missing      int `json:"missing"`
	Unknown      int `json:"unknown"`
	NotReady     int `json:"notReady"`
}

type GitRepoDisplay struct {
	ReadyBundleDeployments string `json:"readyBundleDeployments,omitempty"`
	State                  string `json:"state,omitempty"`
	Message                string `json:"message,omitempty"`
	Error                  bool   `json:"error,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type GitRepoRestriction struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	DefaultServiceAccount  string   `json:"defaultServiceAccount,omitempty"`
	AllowedServiceAccounts []string `json:"allowedServiceAccounts,omitempty"`
	AllowedRepoPatterns    []string `json:"allowedRepoPatterns,omitempty"`

	DefaultClientSecretName  string   `json:"defaultClientSecretName,omitempty"`
	AllowedClientSecretNames []string `json:"allowedClientSecretNames,omitempty"`
}

type GitRepoResource struct {
	APIVersion      string                    `json:"apiVersion,omitempty"`
	Kind            string                    `json:"kind,omitempty"`
	Type            string                    `json:"type,omitempty"`
	ID              string                    `json:"id,omitempty"`
	Namespace       string                    `json:"namespace,omitempty"`
	Name            string                    `json:"name,omitempty"`
	IncompleteState bool                      `json:"incompleteState,omitempty"`
	State           string                    `json:"state,omitempty"`
	Error           bool                      `json:"error,omitempty"`
	Transitioning   bool                      `json:"transitioning,omitempty"`
	Message         string                    `json:"message,omitempty"`
	PerClusterState []ResourcePerClusterState `json:"perClusterState,omitempty"`
}

type ResourcePerClusterState struct {
	State         string      `json:"state,omitempty"`
	Error         bool        `json:"error,omitempty"`
	Transitioning bool        `json:"transitioning,omitempty"`
	Message       string      `json:"message,omitempty"`
	Patch         *GenericMap `json:"patch,omitempty"`
	ClusterID     string      `json:"clusterId,omitempty"`
}