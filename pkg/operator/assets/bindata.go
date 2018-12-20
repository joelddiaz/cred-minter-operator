package assets

import (
	"fmt"
	"strings"
)

var _config_cred_minter_yaml_crds_credminter_v1beta1_credentialsrequest_yaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  creationTimestamp: null
  labels:
    controller-tools.k8s.io: "1.0"
  name: credentialsrequests.credminter.openshift.io
spec:
  group: credminter.openshift.io
  names:
    kind: CredentialsRequest
    plural: credentialsrequests
  scope: Namespaced
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          type: string
        kind:
          type: string
        metadata:
          type: object
        spec:
          properties:
            aws:
              properties:
                statementEntries:
                  items:
                    properties:
                      action:
                        items:
                          type: string
                        type: array
                      effect:
                        type: string
                      resource:
                        type: string
                    required:
                    - effect
                    - action
                    - resource
                    type: object
                  type: array
              required:
              - statementEntries
              type: object
            clusterID:
              type: string
            clusterName:
              type: string
            secretRef:
              type: object
          required:
          - clusterName
          - clusterID
          - secretRef
          type: object
        status:
          properties:
            aws:
              properties:
                user:
                  type: string
              required:
              - user
              type: object
            lastSyncGeneration:
              format: int64
              type: integer
            lastSyncTimestamp:
              format: date-time
              type: string
            provisioned:
              type: boolean
          required:
          - provisioned
          - lastSyncGeneration
          type: object
      required:
      - spec
  version: v1beta1
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`)

func config_cred_minter_yaml_crds_credminter_v1beta1_credentialsrequest_yaml() ([]byte, error) {
	return _config_cred_minter_yaml_crds_credminter_v1beta1_credentialsrequest_yaml, nil
}

var _config_cred_minter_yaml_rbac_rbac_role_yaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  creationTimestamp: null
  name: cred-minter-role
rules:
- apiGroups:
  - credminter.openshift.io
  resources:
  - credentialsrequests
  - credentialsrequests/status
  - credentialsrequests/finalizers
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
`)

func config_cred_minter_yaml_rbac_rbac_role_yaml() ([]byte, error) {
	return _config_cred_minter_yaml_rbac_rbac_role_yaml, nil
}

var _config_cred_minter_yaml_rbac_rbac_role_binding_yaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  creationTimestamp: null
  name: cred-minter-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cred-minter-role
subjects:
- kind: ServiceAccount
  name: default
  namespace: openshift-cred-minter
`)

func config_cred_minter_yaml_rbac_rbac_role_binding_yaml() ([]byte, error) {
	return _config_cred_minter_yaml_rbac_rbac_role_binding_yaml, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		return f()
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() ([]byte, error){
	"config/cred-minter-yaml/crds/credminter_v1beta1_credentialsrequest.yaml": config_cred_minter_yaml_crds_credminter_v1beta1_credentialsrequest_yaml,
	"config/cred-minter-yaml/rbac/rbac_role.yaml":                             config_cred_minter_yaml_rbac_rbac_role_yaml,
	"config/cred-minter-yaml/rbac/rbac_role_binding.yaml":                     config_cred_minter_yaml_rbac_rbac_role_binding_yaml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func     func() ([]byte, error)
	Children map[string]*_bintree_t
}

var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"config": {nil, map[string]*_bintree_t{
		"cred-minter-yaml": {nil, map[string]*_bintree_t{
			"crds": {nil, map[string]*_bintree_t{
				"credminter_v1beta1_credentialsrequest.yaml": {config_cred_minter_yaml_crds_credminter_v1beta1_credentialsrequest_yaml, map[string]*_bintree_t{}},
			}},
			"rbac": {nil, map[string]*_bintree_t{
				"rbac_role.yaml":         {config_cred_minter_yaml_rbac_rbac_role_yaml, map[string]*_bintree_t{}},
				"rbac_role_binding.yaml": {config_cred_minter_yaml_rbac_rbac_role_binding_yaml, map[string]*_bintree_t{}},
			}},
		}},
	}},
}}
